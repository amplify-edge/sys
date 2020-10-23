package accountpkg

import (
	"context"
	"fmt"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/repo"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

const (
	errInvalidConfig = "error validating provided config, %s is %s"
	errRunningServer = "error running grpc & grpc web service: %v"
)

type SysAccountService struct {
	authInterceptorFunc func(context.Context) (context.Context, error)
	proxyService        *pkg.SysAccountProxyService
	dbProxyService      *coresvc.SysCoreProxyService
}

type SysAccountServiceConfig struct {
	store  *coredb.CoreDB
	Cfg    *service.SysAccountConfig
	logger *logrus.Entry
}

func NewSysAccountServiceConfig(l *logrus.Entry, db *coredb.CoreDB, filepath string) (*SysAccountServiceConfig, error) {
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
		store:  db,
		Cfg:    accountCfg,
		logger: sysAccountLogger,
	}
	return sasc, nil
}

func NewSysAccountService(cfg *SysAccountServiceConfig) (*SysAccountService, error) {
	cfg.logger.Infoln("Initializing Sys-Account Service")

	authRepo, err := repo.NewAuthRepo(cfg.logger, cfg.store, cfg.Cfg)
	if err != nil {
		return nil, err
	}
	sysAccountProxy := pkg.NewSysAccountProxyService(authRepo, authRepo)
	dbProxyService := coresvc.NewSysCoreProxyService(cfg.store)
	return &SysAccountService{
		authInterceptorFunc: authRepo.DefaultInterceptor,
		proxyService:        sysAccountProxy,
		dbProxyService:      dbProxyService,
	}, nil
}

func (sas *SysAccountService) InjectInterceptors(unaryItc []grpc.UnaryServerInterceptor, streamItc []grpc.StreamServerInterceptor) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	unaryItc = append(unaryItc, grpcAuth.UnaryServerInterceptor(sas.authInterceptorFunc))
	streamItc = append(streamItc, grpcAuth.StreamServerInterceptor(sas.authInterceptorFunc))
	return unaryItc, streamItc
}

func (sas *SysAccountService) RegisterServices(srv *grpc.Server) {
	sas.proxyService.RegisterSvc(srv)
	sas.dbProxyService.RegisterSvc(srv)
}
