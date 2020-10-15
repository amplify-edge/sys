package accountpkg

import (
	"context"
	"fmt"
	"github.com/genjidb/genji"
	"github.com/getcouragenow/sys/sys-account/service/go"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/repo"
	sysAccountUtil "github.com/getcouragenow/sys/sys-account/service/go/pkg/utilities"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
)

const (
	errInvalidConfig = "error validating provided config, %s is %s"
	errRunningServer = "error running grpc & grpc web service: %v"
)

type SysAccountService struct {
	authInterceptorFunc func(context.Context) (context.Context, error)
	proxyService        *pkg.SysAccountProxyService
}

type SysAccountServiceConfig struct {
	store  *genji.DB
	Cfg    *service.SysAccountConfig
	logger *logrus.Entry
}

func NewSysAccountServiceConfig(l *logrus.Entry, db *genji.DB, unauthenticatedRoutes []string) (*SysAccountServiceConfig, error) {
	var err error
	if db == nil {
		db, err = coredb.SharedDatabase()
		if err != nil {
			return nil, err
		}
	}
	sysAccountLogger := l.WithFields(logrus.Fields{
		"sys": "sys-account",
	})

	sasc := &SysAccountServiceConfig{
		store:  db,
		Cfg:    &service.SysAccountConfig{UnauthenticatedRoutes: unauthenticatedRoutes},
		logger: sysAccountLogger,
	}
	if err := sasc.parseAndValidate(); err != nil {
		return nil, err
	}
	return sasc, nil
}

func (ssc *SysAccountServiceConfig) parseAndValidate() error {
	if ssc.Cfg.JWTConfig.Access.Secret == "" {
		accessSecret, err := sysAccountUtil.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		ssc.Cfg.JWTConfig.Access.Secret = string(accessSecret)
	}
	if ssc.Cfg.JWTConfig.Refresh.Secret == "" {
		refreshSecret, err := sysAccountUtil.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		ssc.Cfg.JWTConfig.Refresh.Secret = string(refreshSecret)
	}
	if ssc.Cfg.UnauthenticatedRoutes == nil {
		return fmt.Errorf(errInvalidConfig, "sys_account.unauthenticatedRoutes", "missing")
	}
	return nil
}

func NewSysAccountService(cfg *SysAccountServiceConfig) (*SysAccountService, error) {
	cfg.logger.Infoln("Initializing Sys-Account Service")

	authRepo, err := repo.NewAuthRepo(cfg.logger, cfg.store, cfg.Cfg)
	if err != nil {
		return nil, err
	}
	sysAccountProxy := pkg.NewSysAccountProxyService(authRepo, authRepo)
	return &SysAccountService{
		authInterceptorFunc: authRepo.DefaultInterceptor,
		proxyService:        sysAccountProxy,
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
