package repo

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"time"

	utilities "github.com/getcouragenow/sys-share/sys-core/service/config"
	corepkg "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/pass"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

const (
	banDuration = 1 * time.Hour
)

func (ad *SysAccountRepo) getAndVerifyAccount(ctx context.Context, req *pkg.LoginRequest) (*pkg.Account, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "SysAccount.GetAndVerifyAccount")
	defer sp.Finish()
	qp := &coredb.QueryParams{Params: map[string]interface{}{
		"email": req.Email,
	}}
	acc, err := ad.store.GetAccount(qp)
	if err != nil {
		return nil, err
	}

	if acc.Disabled {
		return nil, fmt.Errorf(sharedAuth.Error{Reason: sharedAuth.ErrAccountDisabled, Err: fmt.Errorf("password mismatch")}.Error())
	}

	matchedPassword, err := pass.VerifyHash(req.Password, acc.Password)
	if err != nil {
		return nil, err
	}
	if !matchedPassword {
		return nil, fmt.Errorf(sharedAuth.Error{Reason: sharedAuth.ErrVerifyPassword, Err: fmt.Errorf("password mismatch")}.Error())
	}

	ad.log.WithFields(map[string]interface{}{"account_id": acc.ID}).Debug("querying user")

	daoRoles, err := ad.store.FetchRoles(acc.ID)
	if err != nil {
		ad.log.Debugf("unable to fetch user roles: %v", err)
		return nil, err
	}
	var pkgRoles []*pkg.UserRoles
	for _, daoRole := range daoRoles {
		pkgRole, err := daoRole.ToPkgRole()
		if err != nil {
			ad.log.Debugf("unable to convert user role to pkg role: %v", err)
			return nil, err
		}
		pkgRoles = append(pkgRoles, pkgRole)
	}
	return acc.ToPkgAccount(pkgRoles, nil)
}

// Register satisfies rpc.Register function on AuthService proto definition
func (ad *SysAccountRepo) Register(ctx context.Context, in *pkg.RegisterRequest) (*pkg.RegisterResponse, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	if in.Password != in.PasswordConfirm {
		return nil, status.Errorf(codes.InvalidArgument, "password mismatch")
	}
	// New user will be assigned GUEST role and no Org / Project for now.
	// TODO @gutterbacon: subject to change.
	newAcc := &pkg.AccountNewRequest{
		Email:    in.Email,
		Password: in.Password,
		Roles:    []*pkg.UserRoles{},
	}
	if in.UserRoles != nil {
		newAcc.Roles = append(newAcc.Roles, in.UserRoles)
	} else {
		newAcc.Roles = append(newAcc.Roles, &pkg.UserRoles{
			Role: 1,
		})
	}
	acc, err := ad.store.InsertFromPkgAccountRequest(newAcc, false)
	if err != nil {
		return &pkg.RegisterResponse{
			Success:     false,
			ErrorReason: err.Error(),
		}, err
	}

	vtoken, _, err := ad.genVerificationToken(&coredb.QueryParams{Params: map[string]interface{}{"email": in.Email}})
	if err != nil {
		return nil, err
	}

	errChan := make(chan error, 1)
	go func() {
		mailContent, err := ad.mailVerifyAccountTpl(acc.Email, vtoken, acc.ID)
		if err != nil {
			ad.log.Debugf("cannot create verify account email: %v", err)
			errChan <- err
			return
		}
		ad.log.Debugf("Email content: %s", string(mailContent))
		resp, err := ad.mail.SendMail(ctx, &corepkg.EmailRequest{
			Subject: fmt.Sprintf("Verify Account %s Register", acc.Email),
			Recipients: map[string]string{
				acc.Email: acc.Email,
			},
			Content: mailContent,
		})
		if err != nil {
			ad.log.Debugf("cannot send verify account email: %v", err)
			errChan <- err
			return
		}
		ad.log.Debugf("Sent Email to %s => %v\n", acc.Email, resp)
		close(errChan)
	}()
	if err = <-errChan; err != nil {
		ad.log.Errorf("Cannot send email: %v", err)
	}
	registeredUserMetrics := ad.bizmetrics.RegisteredUserMetrics
	go func() {
		registeredUserMetrics.Inc()
	}()
	return &pkg.RegisterResponse{
		Success:     true,
		SuccessMsg:  fmt.Sprintf("Successfully created user: %s as Guest", in.Email),
		ErrorReason: "",
		TempUserId:  acc.ID,
		VerifyToken: vtoken,
	}, nil
}

func (ad *SysAccountRepo) genVerificationToken(param *coredb.QueryParams) (string, *dao.Account, error) {
	// TODO @gutterbacon: verification token, replace this with anything else
	// like OTP or anything
	vtoken := utilities.NewID()
	// update user's account table's verification_token
	acc, err := ad.store.GetAccount(param)
	if err != nil {
		return "", nil, err
	}
	acc.VerificationToken = vtoken
	err = ad.store.UpdateAccount(acc)
	if err != nil {
		return "", nil, err
	}
	return acc.VerificationToken, acc, nil
}

func (ad *SysAccountRepo) Login(ctx context.Context, in *pkg.LoginRequest) (*pkg.LoginResponse, error) {
	if in == nil {
		return &pkg.LoginResponse{}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return &pkg.LoginResponse{}, status.Errorf(codes.Internal, "Unable to get user's ip")
	}
	peerIp := peer.Addr.String()
	clientIp, _, err := net.SplitHostPort(peerIp)
	if err != nil {
		return &pkg.LoginResponse{}, status.Errorf(codes.Internal, "Unable to get user's ip and port")
	}
	loginAttempts, err := ad.store.GetLoginAttempt(clientIp)
	if err != nil {
		loginAttempts, err = ad.store.UpsertLoginAttempt(clientIp, in.Email, 0, 0)
		if err != nil {
			return &pkg.LoginResponse{}, status.Errorf(codes.Internal, "Unable to create user's login attempt")
		}
	}
	if loginAttempts.TotalAttempts > 5 && loginAttempts.BanPeriod != 0 && loginAttempts.BanPeriod >= utilities.CurrentTimestamp() {
		return &pkg.LoginResponse{}, status.Errorf(codes.PermissionDenied, "You've failed to submit correct login information too many times, try again in an hour")
	}
	var claimant sharedAuth.Claimant

	u, err := ad.getAndVerifyAccount(ctx, in)
	if err != nil {
		if loginAttempts.TotalAttempts >= 5 {
			loginAttempts, _ = ad.store.UpsertLoginAttempt(loginAttempts.OriginIP, loginAttempts.AccountEmail, loginAttempts.TotalAttempts+1, utilities.CurrentTimestamp()+banDuration.Nanoseconds())
		} else {
			loginAttempts, _ = ad.store.UpsertLoginAttempt(loginAttempts.OriginIP, loginAttempts.AccountEmail, loginAttempts.TotalAttempts+1, 0)
		}
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, err
	}
	claimant = u

	tokenPairs, err := ad.tokenCfg.NewTokenPairs(claimant)
	if err != nil {
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", sharedAuth.Error{Reason: sharedAuth.ErrCreatingToken, Err: err})
	}

	req, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{"id": u.Id}})
	if err != nil {
		return nil, err
	}
	req.LastLogin = utilities.CurrentTimestamp()
	if err = ad.store.UpdateAccount(req); err != nil {
		return nil, err
	}
	errChan := make(chan error, 1)
	go func() {
		payloadBytes, err := coredb.MarshalToBytes(map[string]interface{}{"accessToken": tokenPairs.AccessToken, "refreshToken": tokenPairs.RefreshToken})
		if err != nil {
			ad.log.Debugf("error while marshal onLoginCreateInterceptor payload: %v", err)
			errChan <- err
			return
		}
		resp, err := ad.bus.Broadcast(ctx, &corepkg.EventRequest{
			EventName:   "onLoginCreateInterceptor",
			Initiator:   "sys-account",
			UserId:      u.Id,
			JsonPayload: payloadBytes,
		})
		if err != nil {
			ad.log.Debugf("error while calling onLoginCreateInterceptor: %v", err)
			errChan <- err
			return
		}
		ad.log.Debugf("event response: %v", string(resp.Reply))
		close(errChan)
	}()
	if err = <-errChan; err != nil {
		ad.log.Errorf("cannot call onLoginCreateInterceptor event: %v", err)
		// return nil, err
	}
	// on success, resets login attempts counter
	_, err = ad.store.UpsertLoginAttempt(loginAttempts.OriginIP, loginAttempts.AccountEmail, 0, 0)
	if err != nil {
		return nil, err
	}
	return &pkg.LoginResponse{
		Success:      true,
		AccessToken:  tokenPairs.AccessToken,
		RefreshToken: tokenPairs.RefreshToken,
		LastLogin:    req.LastLogin,
	}, nil
}

func (ad *SysAccountRepo) ForgotPassword(ctx context.Context, in *pkg.ForgotPasswordRequest) (*pkg.ForgotPasswordResponse, error) {
	if in == nil {
		return &pkg.ForgotPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request forgot password endpoint: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	// TODO @gutterbacon: this is where we should send an email to verify the user
	// We could also add this to audit log trail.
	// for now this method is a stub.
	vtoken, acc, err := ad.genVerificationToken(&coredb.QueryParams{Params: map[string]interface{}{"email": in.Email}})
	if err != nil {
		return &pkg.ForgotPasswordResponse{
			Success:                   false,
			SuccessMsg:                "",
			ErrorReason:               err.Error(),
			ForgotPasswordRequestedAt: utilities.CurrentTimestamp(),
		}, err
	}
	ad.log.Debugf("Generated Verification Token for ForgotPassword: %s", vtoken)
	errChan := make(chan error, 1)
	go func() {
		mailContent, err := ad.mailForgotPassword(acc.Email, vtoken)
		if err != nil {
			errChan <- err
			return
		}
		resp, err := ad.mail.SendMail(ctx, &corepkg.EmailRequest{
			Subject: fmt.Sprintf("Reset %s Password", acc.Email),
			Recipients: map[string]string{
				acc.Email: acc.Email,
			},
			Content: mailContent,
		})
		if err != nil {
			errChan <- err
			return
		}
		ad.log.Debugf("Sent Email to %s => %v\n", acc.Email, resp)
		close(errChan)
	}()
	if err = <-errChan; err != nil {
		ad.log.Errorf("Cannot send email: %v", err)
		// return nil, err
	}

	return &pkg.ForgotPasswordResponse{
		Success:                   true,
		SuccessMsg:                "Reset password token sent",
		ForgotPasswordRequestedAt: utilities.CurrentTimestamp(),
	}, nil
}

func (ad *SysAccountRepo) ResetPassword(ctx context.Context, in *pkg.ResetPasswordRequest) (*pkg.ResetPasswordResponse, error) {
	if in == nil {
		return &pkg.ResetPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request reset password endpoint: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	// TODO @gutterbacon: This is where we should send an email to verify the user
	// We could also add this to audit log trail.
	// but for now this method is a stub.
	if in.Password != in.PasswordConfirm {
		return nil, fmt.Errorf(sharedAuth.Error{Reason: sharedAuth.ErrVerifyPassword, Err: fmt.Errorf("password mismatch")}.Error())
	}
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{"email": in.Email}})
	if err != nil {
		ad.log.Debugf("error getting reset password account: %v", err)
		return &pkg.ResetPasswordResponse{
			Success:                  false,
			SuccessMsg:               "",
			ErrorReason:              err.Error(),
			ResetPasswordRequestedAt: utilities.CurrentTimestamp(),
		}, err
	}
	ad.log.Debugf("reset password account: %v", *acc)
	if acc.VerificationToken != in.VerifyToken {
		ad.log.Debugf("mismatch verification token: wanted %s\n got: %s", acc.VerificationToken, in.VerifyToken)
		return &pkg.ResetPasswordResponse{
			Success:                  false,
			SuccessMsg:               "",
			ErrorReason:              "verification token mismatch",
			ResetPasswordRequestedAt: utilities.CurrentTimestamp(),
		}, err
	}
	newPasswd, err := pass.GenHash(in.Password)
	if err != nil {
		return &pkg.ResetPasswordResponse{
			Success:                  false,
			SuccessMsg:               "",
			ErrorReason:              err.Error(),
			ResetPasswordRequestedAt: utilities.CurrentTimestamp(),
		}, err
	}
	acc.Password = newPasswd
	err = ad.store.UpdateAccount(acc)
	if err != nil {
		return &pkg.ResetPasswordResponse{
			Success:                  false,
			SuccessMsg:               "",
			ErrorReason:              err.Error(),
			ResetPasswordRequestedAt: utilities.CurrentTimestamp(),
		}, err
	}
	return &pkg.ResetPasswordResponse{
		Success:                  true,
		SuccessMsg:               "successfully reset password",
		ErrorReason:              "",
		ResetPasswordRequestedAt: utilities.CurrentTimestamp(),
	}, nil
}

func (ad *SysAccountRepo) VerifyAccount(ctx context.Context, in *pkg.VerifyAccountRequest) (*emptypb.Empty, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot verify account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{"id": in.AccountId}})
	if err != nil {
		return nil, err
	}
	if acc.VerificationToken != in.VerifyToken {
		return nil, status.Errorf(codes.InvalidArgument, "cannot verify account: %v", sharedAuth.Error{Reason: sharedAuth.ErrVerificationTokenMismatch})
	}
	acc.Verified = true
	err = ad.store.UpdateAccount(acc)
	if err != nil {
		return nil, err
	}
	verifiedUserMetrics := ad.bizmetrics.VerifiedUserMetrics
	go func() {
		verifiedUserMetrics.Inc()
	}()
	return &emptypb.Empty{}, nil
}

func (ad *SysAccountRepo) RefreshAccessToken(ctx context.Context, in *pkg.RefreshAccessTokenRequest) (*pkg.RefreshAccessTokenResponse, error) {
	if in == nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters}.Error(),
		}, status.Errorf(codes.InvalidArgument, "cannot request new access token: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	claims, err := ad.tokenCfg.ParseTokenStringToClaim(in.RefreshToken, false)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: sharedAuth.Error{Reason: sharedAuth.ErrInvalidToken}.Error(),
		}, status.Errorf(codes.InvalidArgument, "refresh token is invalid: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidToken})
	}
	newAccessToken, err := ad.tokenCfg.RenewAccessToken(&claims)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: sharedAuth.Error{Reason: sharedAuth.ErrCreatingToken}.Error(),
		}, status.Errorf(codes.Internal, "cannot request new access token from claims: %v", err.Error())
	}
	return &pkg.RefreshAccessTokenResponse{
		AccessToken: newAccessToken,
	}, nil
}
