package repo

import (
	l "github.com/sirupsen/logrus"

	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
	corebus "github.com/getcouragenow/sys-share/sys-core/service/go/pkg/bus"
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	corefile "github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/repo"
	coremail "github.com/getcouragenow/sys/sys-core/service/go/pkg/mailer"
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
		bus                   *corebus.CoreBus
		mail                  *coremail.MailSvc
		frepo                 *corefile.SysFileRepo
	}
)

func NewAuthRepo(l *l.Entry, db *coredb.CoreDB, cfg *service.SysAccountConfig, bus *corebus.CoreBus, mail *coremail.MailSvc, frepo *corefile.SysFileRepo) (*SysAccountRepo, error) {
	accdb, err := dao.NewAccountDB(db, l)
	if err != nil {
		l.Errorf("Error while initializing DAO: %v", err)
		return nil, err
	}
	tokenCfg := sharedAuth.NewTokenConfig([]byte(cfg.SysAccountConfig.JWTConfig.Access.Secret), []byte(cfg.SysAccountConfig.JWTConfig.Refresh.Secret))
	// Register Bus Dispatchers
	repo := &SysAccountRepo{
		store:                 accdb,
		log:                   l,
		tokenCfg:              tokenCfg,
		unauthenticatedRoutes: cfg.SysAccountConfig.UnauthenticatedRoutes,
		bus:                   bus,
		mail:                  mail,
		frepo:                 frepo,
	}
	bus.RegisterAction("onDeleteOrg", repo.onDeleteOrg)
	bus.RegisterAction("onDeleteAccount", repo.onDeleteAccount)
	bus.RegisterAction("onDeleteProject", repo.onDeleteProject)
	bus.RegisterAction("onCheckProjectExists", repo.onCheckProjectExists)
	bus.RegisterAction("onCheckAccountExists", repo.onCheckAccountExists)
	return repo, nil
}
