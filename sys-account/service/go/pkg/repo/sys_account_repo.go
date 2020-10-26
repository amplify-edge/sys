package repo

import (
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	l "github.com/sirupsen/logrus"
)

type (
	// SysAccountRepo is the repository layer of the authn & authz && accounts
	SysAccountRepo struct {
		store    *dao.AccountDB
		log      *l.Entry
		tokenCfg *sharedAuth.TokenConfig
		// the auth interceptor would not intercept tokens on these routes
		// (format is: /ProtoServiceName/ProtoServiceMethod, example: /proto.AuthService/Login).
		unauthenticatedRoutes []string
	}
)

func NewAuthRepo(l *l.Entry, db *coredb.CoreDB, cfg *service.SysAccountConfig) (*SysAccountRepo, error) {
	accdb, err := dao.NewAccountDB(db)
	if err != nil {
		l.Errorf("Error while initializing DAO: %v", err)
		return nil, err
	}
	tokenCfg := sharedAuth.NewTokenConfig([]byte(cfg.SysAccountConfig.JWTConfig.Access.Secret), []byte(cfg.SysAccountConfig.JWTConfig.Refresh.Secret))
	return &SysAccountRepo{
		store:                 accdb,
		log:                   l,
		tokenCfg:              tokenCfg,
		unauthenticatedRoutes: cfg.SysAccountConfig.UnauthenticatedRoutes,
	}, nil
}
