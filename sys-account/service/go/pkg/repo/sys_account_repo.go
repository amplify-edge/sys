package repo

import (
	"github.com/amplify-cms/sys-share/sys-core/service/logging"
	"github.com/amplify-cms/sys/sys-account/service/go/pkg/superusers"
	"github.com/amplify-cms/sys/sys-account/service/go/pkg/telemetry"

	sharedAuth "github.com/amplify-cms/sys-share/sys-account/service/go/pkg/shared"
	corebus "github.com/amplify-cms/sys-share/sys-core/service/go/pkg/bus"
	"github.com/amplify-cms/sys/sys-account/service/go"
	"github.com/amplify-cms/sys/sys-account/service/go/pkg/dao"
	"github.com/amplify-cms/sys/sys-core/service/go/pkg/coredb"
	corefile "github.com/amplify-cms/sys/sys-core/service/go/pkg/filesvc/repo"
	coremail "github.com/amplify-cms/sys/sys-core/service/go/pkg/mailer"
)

type (
	// SysAccountRepo is the repository layer of the authn & authz && accounts
	SysAccountRepo struct {
		store    *dao.AccountDB
		log      logging.Logger
		tokenCfg *sharedAuth.TokenConfig
		// the auth interceptor would not intercept tokens on these routes
		// (format is: /ProtoServiceName/ProtoServiceMethod, example: /proto.AuthService/Login).
		unauthenticatedRoutes []string
		bus                   *corebus.CoreBus
		mail                  *coremail.MailSvc
		frepo                 *corefile.SysFileRepo
		domain                string
		//initialSuperusersMail []string
		bizmetrics *telemetry.SysAccountMetrics
		superDao   *superusers.SuperUserIO
	}
)

func NewAuthRepo(l logging.Logger, db *coredb.CoreDB, cfg *service.SysAccountConfig, bus *corebus.CoreBus, mail *coremail.MailSvc, frepo *corefile.SysFileRepo, domain string, superUserFilePath string, bizmetrics *telemetry.SysAccountMetrics) (*SysAccountRepo, error) {
	accdb, err := dao.NewAccountDB(db, l)
	if err != nil {
		l.Errorf("Error while initializing DAO: %v", err)
		return nil, err
	}
	tokenCfg := sharedAuth.NewTokenConfig([]byte(cfg.SysAccountConfig.JWTConfig.Access.Secret), []byte(cfg.SysAccountConfig.JWTConfig.Refresh.Secret))
	//var initialSuperMails []string
	//for _, sureq := range supes {
	//	initialSuperMails = append(initialSuperMails, sureq.Email)
	//}
	superDao := superusers.NewSuperUserDAO(superUserFilePath, l)
	repo := &SysAccountRepo{
		store:                 accdb,
		log:                   l,
		tokenCfg:              tokenCfg,
		unauthenticatedRoutes: cfg.SysAccountConfig.UnauthenticatedRoutes,
		bus:                   bus,
		mail:                  mail,
		frepo:                 frepo,
		domain:                domain,
		//initialSuperusersMail: initialSuperMails,
		bizmetrics: bizmetrics,
		superDao:   superDao,
	}
	// Register Bus Dispatchers
	bus.RegisterAction("onDeleteOrg", repo.onDeleteOrg)
	bus.RegisterAction("onDeleteAccount", repo.onDeleteAccount)
	bus.RegisterAction("onDeleteProject", repo.onDeleteProject)
	bus.RegisterAction("onCheckProjectExists", repo.onCheckProjectExists)
	bus.RegisterAction("onCheckAccountExists", repo.onCheckAccountExists)
	bus.RegisterAction("onResetAllSysAccount", repo.onResetAllSysAccount)
	bus.RegisterAction("onGetAccountEmail", repo.onGetAccountEmail)
	return repo, nil
}
