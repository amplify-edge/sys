package pkg

import (
	"context"
	"fmt"

	"github.com/genjidb/genji"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"

	"net/http"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	sysAccountServer "github.com/getcouragenow/sys/sys-account/service/go"
	sysAccountDeli "github.com/getcouragenow/sys/sys-account/service/go/pkg/repo"
	sysAccountUtil "github.com/getcouragenow/sys/sys-account/service/go/pkg/utilities"
)

const (
	errInvalidConfig = "error validating provided config, %s is %s"
	errRunningServer = "error running grpc & grpc web service: %v"
)

// SysServices will be the struct provided to the callee of this package
// it contains all sub grpc services contained within the `sys` repo.
// for example it will be:
// - sys-account (auth and account service)
// - sys-core (not sure about db)
// TODO @gutterbacon : When other sys-* are built, put it on sys-share as a proxy, then call it here.
type SysServices struct {
	logger              *logrus.Entry
	authInterceptorFunc func(context.Context) (context.Context, error)
	port                int
	ProxyService        *pkg.SysAccountProxyService
}

// SysServiceConfig contains all the configuration
// for each services, because SysService needs this in order to
// load up and provide sub grpc services.
// TODO @gutterbacon : When other sys-* are built, put it on sys-share as a proxy then call it here.
type SysServiceConfig struct {
	DB         *genji.DB // sys-core
	SysAccount *sysAccountServer.SysAccountConfig
	Port       int
	Logger     *logrus.Entry
}

// TODO @gutterbacon: this function is a stub, we need to load up config from somewhere later.
func NewSysServiceConfig(l *logrus.Entry, db *genji.DB, unauthenticatedRoutes []string, port int) (*SysServiceConfig, error) {
	if db == nil {
		db, _ = coredb.SharedDatabase()
	}
	ssc := &SysServiceConfig{
		Logger:     l,
		DB:         db,
		Port:       port,
		SysAccount: &sysAccountServer.SysAccountConfig{UnauthenticatedRoutes: unauthenticatedRoutes},
	}
	if err := ssc.parseAndValidate(); err != nil {
		return nil, err
	}
	return ssc, nil
}

func (ssc *SysServiceConfig) parseAndValidate() error {
	if ssc.SysAccount.JWTConfig.Access.Secret == "" {
		accessSecret, err := sysAccountUtil.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		ssc.SysAccount.JWTConfig.Access.Secret = string(accessSecret)
	}
	if ssc.SysAccount.JWTConfig.Refresh.Secret == "" {
		refreshSecret, err := sysAccountUtil.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		ssc.SysAccount.JWTConfig.Refresh.Secret = string(refreshSecret)
	}
	if ssc.SysAccount.UnauthenticatedRoutes == nil {
		return fmt.Errorf(errInvalidConfig, "sys_account.unauthenticatedRoutes", "missing")
	}
	return nil
}

// NewService will create new SysServices
// this SysServices could be passed around to other mod-* and maintemplates-*
// or could be run independently using Run method below
func NewService(cfg *SysServiceConfig) (*SysServices, error) {
	// load up the sub grpc Services
	cfg.Logger.Println("Initializing GRPC Services")

	if err := cfg.parseAndValidate(); err != nil {
		return nil, err
	}

	// ========================================================================
	// Sys-Account
	// ========================================================================
	authDeli, err := sysAccountDeli.NewAuthDeli(cfg.Logger, cfg.DB, cfg.SysAccount)
	if err != nil {
		return nil, err
	}

	sysAccountProxy := pkg.NewSysAccountProxyService(authDeli, authDeli)

	// ========================================================================

	return &SysServices{
		logger:              cfg.Logger,
		port:                cfg.Port,
		authInterceptorFunc: authDeli.DefaultInterceptor,
		ProxyService:        sysAccountProxy,
	}, nil
}

// NewServer to the supplied grpc server.
func (s *SysServices) InjectInterceptors(unaryInterceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	recoveryOptions := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(s.recoveryHandler()),
	}
	logrusOpts := []grpcLogrus.Option{
		grpcLogrus.WithLevels(grpcLogrus.DefaultCodeToLevel),
	}

	unaryInterceptors = append(
		unaryInterceptors,
		grpcRecovery.UnaryServerInterceptor(recoveryOptions...),
		grpcLogrus.UnaryServerInterceptor(s.logger, logrusOpts...),
		grpcAuth.UnaryServerInterceptor(s.authInterceptorFunc),
	)

	streamInterceptors = append(
		streamInterceptors,
		grpcRecovery.StreamServerInterceptor(recoveryOptions...),
		grpcLogrus.StreamServerInterceptor(s.logger, logrusOpts...),
		grpcAuth.StreamServerInterceptor(s.authInterceptorFunc),
	)
	return unaryInterceptors, streamInterceptors
}

func (s *SysServices) RegisterServices(srv *grpc.Server) {
	s.ProxyService.RegisterSvc(srv)
}

func (s *SysServices) recoveryHandler() func(panic interface{}) error {
	return func(panic interface{}) error {
		s.logger.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

// Creates a GrpcWeb wrapper around grpc.Server
func (s *SysServices) RegisterGrpcWebServer(srv *grpc.Server) *grpcweb.WrappedGrpcServer {
	return grpcweb.WrapServer(
		srv,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
		grpcweb.WithWebsockets(true),
	)
}

// run runs all the sys-* service as a service
func (s *SysServices) run(grpcWebServer *grpcweb.WrappedGrpcServer, httpServer *http.Server) error {
	if httpServer == nil {
		httpServer = &http.Server{
			Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web")
				s.logger.Infof("Request Endpoint: %s", r.URL)
				grpcWebServer.ServeHTTP(w, r)
			}), &http2.Server{}),
		}
	}

	httpServer.Addr = fmt.Sprintf("127.0.0.1:%d", s.port)
	return httpServer.ListenAndServe()
}

// Run is just an exported wrapper for s.run()
func (s *SysServices) Run(srv *grpcweb.WrappedGrpcServer, httpServer *http.Server) {
	if err := s.run(srv, httpServer); err != nil {
		s.logger.Fatalf(errRunningServer, err)
	}
}
