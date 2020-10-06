package delivery

import (
	"context"
	"fmt"

	"github.com/genjidb/genji"
	server "github.com/getcouragenow/sys/sys-account/service/go"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	l "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/dao"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
)

type (
	// ContextKey auth gRPC context key.
	ContextKey string

	// AuthDelivery is the delivery layer of the authn
	AuthDelivery struct {
		store                 *dao.AccountDB
		Log                   *l.Entry
		TokenCfg              *auth.TokenConfig
		UnauthenticatedRoutes []string // the auth interceptor would not intercept tokens on these routes (format is: /ProtoServiceName/ProtoServiceMethod, example: /proto.AuthService/Login).
		// repo ==> some repository layer for querying users
	}
)

func NewAuthDeli(l *l.Entry, db *genji.DB, cfg *server.SysAccountConfig) (*AuthDelivery, error) {
	accdb, err := dao.NewAccountDB(db)
	if err != nil {
		return nil, err
	}
	tokenCfg := auth.NewTokenConfig([]byte(cfg.JWTConfig.Access.Secret), []byte(cfg.JWTConfig.Refresh.Secret))
	return &AuthDelivery{
		store:                 accdb,
		Log:                   l,
		TokenCfg:              tokenCfg,
		UnauthenticatedRoutes: cfg.UnauthenticatedRoutes,
	}, nil
}

func (c ContextKey) String() string {
	return "sys_account.delivery.grpc context key " + string(c)
}

var (
	// ContextKeyClaims defines claims context key provided by token.
	ContextKeyClaims = ContextKey("auth-claims")
	HeaderAuthorize  = "authorization"
)

// for now we hardcode the user first
// later we'll use Genji from getcouragenow/sys-core/service/db
func (ad *AuthDelivery) getAndVerifyAccount(_ context.Context, req *pkg.LoginRequest) (*pkg.Account, error) {
	qp := &dao.QueryParams{Params: map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
	}}
	acc, err := ad.store.GetAccount(qp)
	if err != nil {
		ad.Log.Warnf("error while querying account: %v", err)
		return nil, err
	}
	ad.Log.WithFields(l.Fields{
		"account_id": acc.ID,
		"role_id":    acc.RoleId,
	}).Info("querying user")
	qp = &dao.QueryParams{Params: map[string]interface{}{"account_id": acc.ID}}
	role, err := ad.store.GetRole(qp)
	if err != nil {
		ad.Log.Warnf("error while querying role: %v", err)
		return nil, err
	}
	userRole, err := role.ToPkgRole()
	if err != nil {
		return nil, err
	}
	return acc.ToPkgAccount(userRole)
}

// DefaultInterceptor is default authN/authZ interceptor, validates only token correctness without performing any role specific authorization.
func (ad *AuthDelivery) DefaultInterceptor(ctx context.Context) (context.Context, error) {
	methodName, ok := grpc.Method(ctx)
	if ok {
		ad.Log.Infof("Method being called: %s", methodName)
	}

	// Simply returns the context if request are being made on unauthenticated service path / method.
	for _, routes := range ad.UnauthenticatedRoutes {
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
func (ad *AuthDelivery) Register(ctx context.Context, in *pkg.RegisterRequest) (*pkg.RegisterResponse, error) {
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

func (ad *AuthDelivery) Login(ctx context.Context, in *pkg.LoginRequest) (*pkg.LoginResponse, error) {
	if in == nil {
		return &pkg.LoginResponse{}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", auth.AuthError{Reason: auth.ErrInvalidParameters})
	}
	var claimant auth.Claimant

	u, err := ad.getAndVerifyAccount(ctx, in)
	if err != nil {
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, err
	}
	claimant = u

	tokenPairs, err := ad.TokenCfg.NewTokenPairs(claimant)
	if err != nil {
		return &pkg.LoginResponse{
			ErrorReason: err.Error(),
		}, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", auth.AuthError{Reason: auth.ErrCreatingToken, Err: err})
	}
	return &pkg.LoginResponse{
		Success:      true,
		AccessToken:  tokenPairs.AccessToken,
		RefreshToken: tokenPairs.RefreshToken,
	}, nil
}

func (ad *AuthDelivery) ForgotPassword(ctx context.Context, in *pkg.ForgotPasswordRequest) (*pkg.ForgotPasswordResponse, error) {
	if in == nil {
		return &pkg.ForgotPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request forgot password endpoint: %v", auth.AuthError{Reason: auth.ErrInvalidParameters})
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

func (ad *AuthDelivery) ResetPassword(ctx context.Context, in *pkg.ResetPasswordRequest) (*pkg.ResetPasswordResponse, error) {
	if in == nil {
		return &pkg.ResetPasswordResponse{}, status.Errorf(codes.InvalidArgument, "cannot request reset password endpoint: %v", auth.AuthError{Reason: auth.ErrInvalidParameters})
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

func (ad *AuthDelivery) RefreshAccessToken(ctx context.Context, in *pkg.RefreshAccessTokenRequest) (*pkg.RefreshAccessTokenResponse, error) {
	if in == nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.AuthError{Reason: auth.ErrInvalidParameters}.Error(),
		}, status.Errorf(codes.InvalidArgument, "cannot request new access token: %v", auth.AuthError{Reason: auth.ErrInvalidParameters})
	}
	claims, err := ad.TokenCfg.ParseTokenStringToClaim(in.RefreshToken, false)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.AuthError{Reason: auth.ErrInvalidToken}.Error(),
		}, status.Errorf(codes.InvalidArgument, "refresh token is invalid: %v", auth.AuthError{Reason: auth.ErrInvalidToken})
	}
	newAccessToken, err := ad.TokenCfg.RenewAccessToken(&claims)
	if err != nil {
		return &pkg.RefreshAccessTokenResponse{
			ErrorReason: auth.AuthError{Reason: auth.ErrCreatingToken}.Error(),
		}, status.Errorf(codes.Internal, "cannot request new access token from claims: %v", err.Error())
	}
	return &pkg.RefreshAccessTokenResponse{
		AccessToken: newAccessToken,
	}, nil
}

// ObtainAccessClaimsFromMetadata obtains token claims from given context with gRPC metadata.
func (ad *AuthDelivery) ObtainAccessClaimsFromMetadata(ctx context.Context, isAccess bool) (claims auth.TokenClaims, err error) {
	var authmeta string
	if authmeta, err = ad.fromMetadata(ctx); err != nil {
		return auth.TokenClaims{}, err
	}

	if claims, err = ad.TokenCfg.ParseTokenStringToClaim(authmeta, isAccess); err != nil {
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

func (ad *AuthDelivery) fromMetadata(ctx context.Context) (authMeta string, err error) {
	authMeta = metautils.ExtractIncoming(ctx).Get(HeaderAuthorize)
	if authMeta == "" {
		return "", auth.AuthError{Reason: auth.ErrMissingToken}
	}
	return authMeta, nil
}
