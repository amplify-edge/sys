package accountpkg

import (
	"context"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging"
	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/telemetry"
	"google.golang.org/grpc"

	rpc "go.amplifyedge.org/sys-share-v2/sys-account/service/go/rpc/v2"
	sharedBus "go.amplifyedge.org/sys-share-v2/sys-core/service/go/pkg/bus"
	coreRpc "go.amplifyedge.org/sys-share-v2/sys-core/service/go/rpc/v2"
	"go.amplifyedge.org/sys-v2/sys-account/service/go"
	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/repo"
	"go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/coredb"
	corefile "go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/filesvc/repo"
	coremail "go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/mailer"
)

type SysAccountService struct {
	authInterceptorFunc func(context.Context) (context.Context, error)
	DbService           coreRpc.DbAdminServiceServer
	BusService          coreRpc.BusServiceServer
	MailService         coreRpc.EmailServiceServer
	AuthRepo            *repo.SysAccountRepo
	BusinessTelemetry   *telemetry.SysAccountMetrics
	AllDBs              *coredb.AllDBService
}

type SysAccountServiceConfig struct {
	store    *coredb.CoreDB
	Cfg      *service.SysAccountConfig
	bus      *sharedBus.CoreBus
	mail     *coremail.MailSvc
	logger   logging.Logger
	fileRepo *corefile.SysFileRepo
	allDbs   *coredb.AllDBService
}

func NewSysAccountServiceConfig(l logging.Logger, filepath string, bus *sharedBus.CoreBus, accountCfg *service.SysAccountConfig) (*SysAccountServiceConfig, error) {
	var err error
	sysAccountLogger := l.WithFields(map[string]interface{}{"service": "sys-account"})

	if filepath != "" {
		accountCfg, err = service.NewConfig(filepath)
		if err != nil {
			return nil, err
		}
	}
	// accounts database
	db, err := coredb.NewCoreDB(l, &accountCfg.SysCoreConfig, nil)
	if err != nil {
		return nil, err
	}

	mailSvc := coremail.NewMailSvc(&accountCfg.MailConfig, l)
	// files database
	fileDb, err := coredb.NewCoreDB(l, &accountCfg.SysFileConfig, nil)
	if err != nil {
		return nil, err
	}
	fileRepo, err := corefile.NewSysFileRepo(fileDb, l)
	if err != nil {
		return nil, err
	}

	allDb := coredb.NewAllDBService()
	sysAccountLogger.Info("registering sys-accounts db & filedb to allDb service")
	allDb.RegisterCoreDB(db)
	allDb.RegisterCoreDB(fileDb)

	sasc := &SysAccountServiceConfig{
		store:    db,
		Cfg:      accountCfg,
		bus:      bus,
		logger:   sysAccountLogger,
		mail:     mailSvc,
		fileRepo: fileRepo,
		allDbs:   allDb,
	}
	return sasc, nil
}

func NewSysAccountService(cfg *SysAccountServiceConfig, domain string) (*SysAccountService, error) {
	cfg.logger.Info("Initializing Sys-Account Service")

	sysAccountMetrics := telemetry.NewSysAccountMetrics(cfg.logger)

	authRepo, err := repo.NewAuthRepo(cfg.logger, cfg.store, cfg.Cfg, cfg.bus, cfg.mail, cfg.fileRepo, domain,
		cfg.Cfg.SuperUserFilePath, sysAccountMetrics)
	if err != nil {
		return nil, err
	}

	dbService := cfg.allDbs
	busService := cfg.bus
	mailSvc := cfg.mail

	return &SysAccountService{
		authInterceptorFunc: authRepo.DefaultInterceptor,
		AuthRepo:            authRepo,
		DbService:           dbService,
		BusService:          busService,
		MailService:         mailSvc,
		BusinessTelemetry:   sysAccountMetrics,
		AllDBs:              cfg.allDbs,
	}, nil
}

func (sas *SysAccountService) InjectInterceptors(unaryItc []grpc.UnaryServerInterceptor, streamItc []grpc.StreamServerInterceptor) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	unaryItc = append(unaryItc, grpcAuth.UnaryServerInterceptor(sas.authInterceptorFunc))
	streamItc = append(streamItc, grpcAuth.StreamServerInterceptor(sas.authInterceptorFunc))
	return unaryItc, streamItc
}

func (sas *SysAccountService) RegisterServices(srv *grpc.Server) {
	rpc.RegisterAccountServiceServer(srv, sas.AuthRepo)
	rpc.RegisterAuthServiceServer(srv, sas.AuthRepo)
	rpc.RegisterOrgProjServiceServer(srv, sas.AuthRepo)
	coreRpc.RegisterDbAdminServiceServer(srv, sas.DbService)
	coreRpc.RegisterBusServiceServer(srv, sas.BusService)
	coreRpc.RegisterEmailServiceServer(srv, sas.MailService)
}
