package repo

import (
	"context"
	"fmt"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/pass"

	l "github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
)

// for now we hardcode the user first
// later we'll use Genji from getcouragenow/sys-core/service/db
func (ad *SysAccountRepo) getAndVerifyAccount(_ context.Context, req *pkg.LoginRequest) (*pkg.Account, error) {
	qp := &dao.QueryParams{Params: map[string]interface{}{
		"email": req.Email,
	}}
	acc, err := ad.store.GetAccount(qp)
	if err != nil {
		ad.log.Warnf(auth.Error{Reason: auth.ErrQueryAccount, Err: err}.Error())
		return nil, err
	}
	matchedPassword, err := pass.VerifyHash(req.Password, acc.Password)
	if err != nil {
		ad.log.Warnf(auth.Error{Reason: auth.ErrVerifyPassword, Err: err}.Error())
		return nil, err
	}
	if !matchedPassword {
		ad.log.Warnf(auth.Error{Reason: auth.ErrVerifyPassword, Err: fmt.Errorf("password mismatch")}.Error())
		return nil, err
	}
	ad.log.WithFields(l.Fields{
		"account_id": acc.ID,
		"role_id":    acc.RoleId,
	}).Info("querying user")
	qp = &dao.QueryParams{Params: map[string]interface{}{"account_id": acc.ID}}
	role, err := ad.store.GetRole(qp)
	if err != nil {
		ad.log.Warnf(auth.Error{Reason: auth.ErrQueryAccount, Err: err}.Error())
		return nil, err
	}
	userRole, err := role.ToPkgRole()
	if err != nil {
		return nil, err
	}
	return acc.ToPkgAccount(userRole)
}

// DefaultInterceptor is default authN/authZ interceptor, validates only token correctness without performing any role specific authorization.
func (ad *SysAccountRepo) DefaultInterceptor(ctx context.Context) (context.Context, error) {
	methodName, ok := grpc.Method(ctx)
	if ok {
		ad.log.Infof("Method being called: %s", methodName)
	}

	// Simply returns the context if request are being made on unauthenticated service path / method.
	for _, routes := range ad.unauthenticatedRoutes {
		if routes == methodName {
			return ctx, nil
		}
	}

	claims, err := ad.ObtainAccessClaimsFromMetadata(ctx, true)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Request unauthenticated with error: %v", err)
	}

	return context.WithValue(ctx, ContextKeyClaims, claims), nil
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
	roleId := coredb.UID()
	accountId := coredb.UID()
	now := timestampNow()
	err := ad.store.InsertAccount(&dao.Account{
		ID:                accountId,
		Email:             in.Email,
		Password:          in.Password,
		RoleId:            roleId,
		CreatedAt:         now,
		UserDefinedFields: map[string]interface{}{},
		Disabled:          false,
	})
	if err != nil {
		return &pkg.RegisterResponse{
			Success:     false,
			ErrorReason: err.Error(),
		}, err
	}
	err = ad.store.InsertRole(&dao.Permission{
		ID:        roleId,
		AccountId: accountId,
		Role:      1,
		CreatedAt: now,
	})
	if err != nil {
		return &pkg.RegisterResponse{
			Success:     false,
			ErrorReason: err.Error(),
		}, err
	}
	return &pkg.RegisterResponse{
		Success:     true,
		SuccessMsg:  fmt.Sprintf("Successfully created user: %s as Guest", in.Email),
		ErrorReason: "",
	}, nil
}

func (ad *SysAccountRepo) Login(ctx context.Context, in *pkg.LoginRequest) (*pkg.LoginResponse, error) {
	if in == nil {
		return &pkg.LoginResponse{}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	var claimant auth.Claimant

	u, err := ad.getAndVerifyAccount(ctx, in)
	if err != nil {
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, err
	}
	claimant = u

	tokenPairs, err := ad.tokenCfg.NewTokenPairs(claimant)
	if err != nil {
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", auth.Error{Reason: auth.ErrCreatingToken, Err: err})
	}
	return &pkg.LoginResponse{
		Success:      true,
		AccessToken:  tokenPairs.AccessToken,
		RefreshToken: tokenPairs.RefreshToken,
	}, nil
}

func (ad *SysAccountRepo) ForgotPassword(ctx context.Context, in *pkg.ForgotPasswordRequest) (*pkg.ForgotPasswordResponse, error) {
	if in == nil {
		return &pkg.ForgotPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request forgot password endpoint: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	// TODO @gutterbacon: this is where we should send an email to verify the user
	// We could also add this to audit log trail.
	// for now this method is a stub.
	return &pkg.ForgotPasswordResponse{
		Success:                   false,
		ErrorReason:               "Unimplemented method",
		ForgotPasswordRequestedAt: timestampNow(),
	}, nil
}

func (ad *SysAccountRepo) ResetPassword(ctx context.Context, in *pkg.ResetPasswordRequest) (*pkg.ResetPasswordResponse, error) {
	if in == nil {
		return &pkg.ResetPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request reset password endpoint: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	// TODO @gutterbacon: This is where we should send an email to verify the user
	// We could also add this to audit log trail.
	// but for now this method is a stub.
	return &pkg.ResetPasswordResponse{
		Success:                  false,
		SuccessMsg:               "",
		ErrorReason:              "Unimplemented method",
		ResetPasswordRequestedAt: timestampNow(),
	}, nil
}

func (ad *SysAccountRepo) RefreshAccessToken(ctx context.Context, in *pkg.RefreshAccessTokenRequest) (*pkg.RefreshAccessTokenResponse, error) {
	if in == nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.Error{Reason: auth.ErrInvalidParameters}.Error(),
		}, status.Errorf(codes.InvalidArgument, "cannot request new access token: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	claims, err := ad.tokenCfg.ParseTokenStringToClaim(in.RefreshToken, false)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.Error{Reason: auth.ErrInvalidToken}.Error(),
		}, status.Errorf(codes.InvalidArgument, "refresh token is invalid: %v", auth.Error{Reason: auth.ErrInvalidToken})
	}
	newAccessToken, err := ad.tokenCfg.RenewAccessToken(&claims)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.Error{Reason: auth.ErrCreatingToken}.Error(),
		}, status.Errorf(codes.Internal, "cannot request new access token from claims: %v", err.Error())
	}
	return &pkg.RefreshAccessTokenResponse{
		AccessToken: newAccessToken,
	}, nil
}

// ObtainAccessClaimsFromMetadata obtains token claims from given context with gRPC metadata.
func (ad *SysAccountRepo) ObtainAccessClaimsFromMetadata(ctx context.Context, isAccess bool) (claims auth.TokenClaims, err error) {
	var authmeta string
	if authmeta, err = ad.fromMetadata(ctx); err != nil {
		return auth.TokenClaims{}, err
	}

	if claims, err = ad.tokenCfg.ParseTokenStringToClaim(authmeta, isAccess); err != nil {
		return auth.TokenClaims{}, err
	}

	return claims, nil
}

// ObtainClaimsFromContext obtains token claims from given context with value.
// TODO @gutterbacon: see ../policy/stub.md
func ObtainClaimsFromContext(ctx context.Context) auth.TokenClaims {
	claims, ok := ctx.Value(ContextKeyClaims).(auth.TokenClaims)
	if !ok {
		return auth.TokenClaims{}
	}
	return claims
}

func (ad *SysAccountRepo) fromMetadata(ctx context.Context) (authMeta string, err error) {
	authMeta = metautils.ExtractIncoming(ctx).Get(HeaderAuthorize)
	if authMeta == "" {
		return "", auth.Error{Reason: auth.ErrMissingToken}
	}
	return authMeta, nil
}
