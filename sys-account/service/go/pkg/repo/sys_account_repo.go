package repo

import (
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
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

func NewAuthRepo(l *l.Entry, db *coredb.CoreDB, cfg *service.SysAccountConfig) (*SysAccountRepo, error) {
	accdb, err := dao.NewAccountDB(db)
	if err != nil {
		l.Errorf("Error while initializing DAO: %v", err)
		return nil, err
	}
	tokenCfg := auth.NewTokenConfig([]byte(cfg.SysAccountConfig.JWTConfig.Access.Secret), []byte(cfg.SysAccountConfig.JWTConfig.Refresh.Secret))
	return &SysAccountRepo{
		store:                 accdb,
		log:                   l,
		tokenCfg:              tokenCfg,
		unauthenticatedRoutes: cfg.SysAccountConfig.UnauthenticatedRoutes,
	}, nil
}
