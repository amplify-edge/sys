package accountpkg

import (
	"context"
	"fmt"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	corefile "github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/repo"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/mailer"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	sharedBus "github.com/getcouragenow/sys-share/sys-core/service/go/pkg/bus"
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/repo"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	coremail "github.com/getcouragenow/sys/sys-core/service/go/pkg/mailer"
)

type SysAccountService struct {
	authInterceptorFunc func(context.Context) (context.Context, error)
	proxyService        *pkg.SysAccountProxyService
	DbProxyService      *coresvc.SysCoreProxyService
	BusProxyService     *coresvc.SysBusProxyService
}

type SysAccountServiceConfig struct {
	store    *coredb.CoreDB
	Cfg      *service.SysAccountConfig
	bus      *sharedBus.CoreBus
	mail     *coremail.MailSvc
	logger   *logrus.Entry
	fileRepo *corefile.SysFileRepo
}

func NewSysAccountServiceConfig(l *logrus.Entry, db *coredb.CoreDB, filepath string, bus *sharedBus.CoreBus, mailSvc *mailer.MailSvc, fileRepo *corefile.SysFileRepo) (*SysAccountServiceConfig, error) {
	var err error
	if db == nil {
		return nil, fmt.Errorf("error creating sys account service: database is null")
	}
	sysAccountLogger := l.WithFields(logrus.Fields{
		"sys": "sys-account",
	})

	accountCfg, err := service.NewConfig(filepath)
	if err != nil {
		return nil, err
	}

	sasc := &SysAccountServiceConfig{
		store:    db,
		Cfg:      accountCfg,
		bus:      bus,
		logger:   sysAccountLogger,
		mail:     mailSvc,
		fileRepo: fileRepo,
	}
	return sasc, nil
}

func NewSysAccountService(cfg *SysAccountServiceConfig) (*SysAccountService, error) {
	cfg.logger.Infoln("Initializing Sys-Account Service")

	authRepo, err := repo.NewAuthRepo(cfg.logger, cfg.store, cfg.Cfg, cfg.bus, cfg.mail, cfg.fileRepo)
	if err != nil {
		return nil, err
	}
	sysAccountProxy := pkg.NewSysAccountProxyService(authRepo, authRepo, authRepo)
	dbProxyService := coresvc.NewSysCoreProxyService(cfg.store)
	busProxyService := coresvc.NewSysBusProxyService(cfg.bus)
	for _, users := range cfg.Cfg.SysAccountConfig.InitialSuperUsers {
		err = authRepo.InitSuperUser(&repo.SuperAccountRequest{
			Email:    users.Email,
			Password: users.Password,
		})
		if err != nil {
			return nil, err
		}
	}

	return &SysAccountService{
		authInterceptorFunc: authRepo.DefaultInterceptor,
		proxyService:        sysAccountProxy,
		DbProxyService:      dbProxyService,
		BusProxyService:     busProxyService,
	}, nil
}

func (sas *SysAccountService) InjectInterceptors(unaryItc []grpc.UnaryServerInterceptor, streamItc []grpc.StreamServerInterceptor) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	unaryItc = append(unaryItc, grpcAuth.UnaryServerInterceptor(sas.authInterceptorFunc))
	streamItc = append(streamItc, grpcAuth.StreamServerInterceptor(sas.authInterceptorFunc))
	return unaryItc, streamItc
}

func (sas *SysAccountService) RegisterServices(srv *grpc.Server) {
	sas.proxyService.RegisterSvc(srv)
}
