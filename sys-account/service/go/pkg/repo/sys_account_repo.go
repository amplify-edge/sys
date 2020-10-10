package repo

import (
	"github.com/genjidb/genji"

	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	l "github.com/sirupsen/logrus"
)

type (
	// ContextKey auth gRPC context key.
	ContextKey string

	// SysAccountRepo is the repository layer of the authn & authz && accounts
	SysAccountRepo struct {
		store                 *dao.AccountDB
		log                   *l.Entry
		tokenCfg              *auth.TokenConfig
		unauthenticatedRoutes []string // the auth interceptor would not intercept tokens on these routes (format is: /ProtoServiceName/ProtoServiceMethod, example: /proto.AuthService/Login).
		// repo ==> some repository layer for querying users
	}
)

var (
	// ContextKeyClaims defines claims context key provided by token.
	ContextKeyClaims = ContextKey("auth-claims")
	HeaderAuthorize  = "authorization"
)

func (c ContextKey) String() string {
	return "sys_account.repo.grpc context key " + string(c)
}

func NewAuthRepo(l *l.Entry, db *genji.DB, cfg *service.SysAccountConfig) (*SysAccountRepo, error) {
	accdb, err := dao.NewAccountDB(db)
	if err != nil {
		return nil, err
	}
	tokenCfg := auth.NewTokenConfig([]byte(cfg.JWTConfig.Access.Secret), []byte(cfg.JWTConfig.Refresh.Secret))
	return &SysAccountRepo{
		store:                 accdb,
		log:                   l,
		tokenCfg:              tokenCfg,
		unauthenticatedRoutes: cfg.UnauthenticatedRoutes,
	}, nil
}
