// package pkg provides sub grpc services for mod-* and deploy/templates/maintemplateV*
package pkg

import (
	"context"
	"fmt"
	"github.com/genjidb/genji"

	"net/http"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcMw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"github.com/getcouragenow/sys-share/pkg"

	sysAccountServer "github.com/getcouragenow/sys/sys-account/service/go"
	sysAccountDeli "github.com/getcouragenow/sys/sys-account/service/go/delivery"
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
// TODO @gutterbacon : When other sys-* are built, put it here.
type SysServices struct {
	logger              *logrus.Entry
	authInterceptorFunc func(context.Context) (context.Context, error)
	ProxyService        *pkg.SysShareProxyService
}

// SysServiceConfig contains all the configuration
// for each services, because SysService needs this in order to
// load up and provide sub grpc services.
// TODO @gutterbacon : When other sys-* are built, put it here.
type SysServiceConfig struct {
	DB         *genji.DB // sys-core
	SysAccount *sysAccountServer.SysAccountConfig
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
	log := logrus.New().WithField("sys-pkg", "sys-services")
	// load up the sub grpc Services
	log.Println("Initializing GRPC Services")

	if err := cfg.parseAndValidate(); err != nil {
		return nil, err
	}

	// ========================================================================
	// Sys-Account
	// ========================================================================
	authDeli, err := sysAccountDeli.NewAuthDeli(log, cfg.DB, cfg.SysAccount)
	if err != nil {
		return nil, err
	}

	sysAccountProxy := pkg.NewSysShareProxyService(authDeli, authDeli)

	// ========================================================================

	return &SysServices{
		logger:              log,
		authInterceptorFunc: authDeli.DefaultInterceptor,
		ProxyService:        sysAccountProxy,
	}, nil
}

// RegisterServices method's job is to check if all the services
// contained within SysServices is valid
// if yes, it will register the service to the provided grpcServer
// if not, then it won't.
// For instance:
// SysServices {
// 		AuthService: exists
//      AccountService: nil
// }
// then it will only register AuthService to the grpc.Server
// the method returns the provided grpc.Server if it exists
// or create one if it's not.
func (s *SysServices) RegisterServices(srv *grpc.Server) *grpc.Server {
	if srv == nil {
		recoveryOptions := []grpcRecovery.Option{
			grpcRecovery.WithRecoveryHandler(s.recoveryHandler()),
		}

		logrusOpts := []grpcLogrus.Option{
			grpcLogrus.WithLevels(grpcLogrus.DefaultCodeToLevel),
		}

		srv = grpc.NewServer(
			grpcMw.WithUnaryServerChain(
				grpcRecovery.UnaryServerInterceptor(recoveryOptions...),
				grpcLogrus.UnaryServerInterceptor(s.logger, logrusOpts...),
				grpcAuth.UnaryServerInterceptor(s.authInterceptorFunc),
			),
			grpcMw.WithStreamServerChain(
				grpcRecovery.StreamServerInterceptor(recoveryOptions...),
				grpcLogrus.StreamServerInterceptor(s.logger, logrusOpts...),
				grpcAuth.StreamServerInterceptor(s.authInterceptorFunc),
			),
		)
	}

	s.ProxyService.RegisterSvc(srv)
	return srv
}

func (s *SysServices) recoveryHandler() func(panic interface{}) error {
	return func(panic interface{}) error {
		s.logger.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

// run runs all the sys-* service as a service
func (s *SysServices) run() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	grpcSrv := s.RegisterServices(nil)
	reflection.Register(grpcSrv)

	grpcWebServer := grpcweb.WrapServer(
		grpcSrv,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
		grpcweb.WithWebsockets(true),
	)

	httpServer := &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web")
			s.logger.Infof("Request Endpoint: %s", r.URL)
			grpcWebServer.ServeHTTP(w, r)
		}), &http2.Server{}),
	}

	return httpServer.ListenAndServe()
}

// Run is just an exported wrapper for s.run()
func (s *SysServices) Run() {
	if err := s.run(); err != nil {
		s.logger.Fatalf(errRunningServer, err)
	}
}
