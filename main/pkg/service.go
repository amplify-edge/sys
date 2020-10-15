package pkg

import (
	"fmt"
	"net/http"

	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/genjidb/genji"
	accountpkg "github.com/getcouragenow/sys/sys-account/service/go/pkg"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
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
	logger        *logrus.Entry
	port          int
	sysAccountSvc *accountpkg.SysAccountService
}

// SysServiceConfig contains all the configuration
// for each services, because SysService needs this in order to
// load up and provide sub grpc services.
// TODO @gutterbacon : When other sys-* are built, put it on sys-share as a proxy then call it here.
type SysServiceConfig struct {
	store      *genji.DB // sys-core
	SysAccount *accountpkg.SysAccountServiceConfig
	port       int
	logger     *logrus.Entry
}

// TODO @gutterbacon: this function is a stub, we need to load up config from somewhere later.
func NewSysServiceConfig(l *logrus.Entry, db *genji.DB, unauthenticatedRoutes []string, port int) (*SysServiceConfig, error) {
	var err error
	if db == nil {
		db, err = coredb.SharedDatabase()
		if err != nil {
			return nil, err
		}
	}
	newSysAccountCfg, err := accountpkg.NewSysAccountServiceConfig(l, db, unauthenticatedRoutes)
	if err != nil {
		return nil, err
	}
	ssc := &SysServiceConfig{
		logger:     l,
		store:      db,
		port:       port,
		SysAccount: newSysAccountCfg,
	}
	return ssc, nil
}

// NewService will create new SysServices
// this SysServices could be passed around to other mod-* and maintemplates-*
// or could be run independently using Run method below
func NewService(cfg *SysServiceConfig) (*SysServices, error) {
	// load up the sub grpc Services
	cfg.logger.Println("Initializing GRPC Services")

	// ========================================================================
	// Sys-Account
	// ========================================================================
	sysAccountSvc, err := accountpkg.NewSysAccountService(cfg.SysAccount)
	if err != nil {
		return nil, err
	}
	// ========================================================================

	return &SysServices{
		logger:        cfg.logger,
		port:          cfg.port,
		sysAccountSvc: sysAccountSvc,
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
	// inject unary interceptors
	unaryInterceptors = append(
		unaryInterceptors,
		grpcRecovery.UnaryServerInterceptor(recoveryOptions...),
		grpcLogrus.UnaryServerInterceptor(s.logger, logrusOpts...),
	)
	// inject stream interceptors
	streamInterceptors = append(
		streamInterceptors,
		grpcRecovery.StreamServerInterceptor(recoveryOptions...),
		grpcLogrus.StreamServerInterceptor(s.logger, logrusOpts...),
	)
	// inject grpc auth
	unaryInterceptors, streamInterceptors = s.sysAccountSvc.InjectInterceptors(unaryInterceptors, streamInterceptors)
	return unaryInterceptors, streamInterceptors
}

func (s *SysServices) RegisterServices(srv *grpc.Server) {
	s.sysAccountSvc.RegisterServices(srv)
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
