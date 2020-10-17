// package main
// this is only used for local testing
// making sure that the sys-account service works locally before wiring it up to the maintemplate.
package main

import (
	"fmt"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg"
	"net/http"

	// external
	"github.com/genjidb/genji"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpcMw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"

	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
)

var (
	gdb                          *genji.DB
	defaultUnauthenticatedRoutes = []string{
		"/v2.services.AuthService/Login",
		"/v2.services.AuthService/Register",
		"/v2.services.AuthService/ResetPassword",
		"/v2.services.AuthService/ForgotPassword",
		"/v2.services.AuthService/RefreshAccessToken",
		// debugging purposes
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}
	defaultAddr = "127.0.0.1"
	defaultPort = 8888
)

func recoveryHandler(l *logrus.Entry) func(panic interface{}) error {
	return func(panic interface{}) error {
		l.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

func initdb() {
	gdb, _ = coredb.SharedDatabase()
}

func main() {
	initdb()
	log := logrus.New().WithField("svc", "sys-account")
	sysAccountConfig, err := accountpkg.NewSysAccountServiceConfig(log, gdb, "path")
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	svc, err := accountpkg.NewSysAccountService(sysAccountConfig)
	if err != nil {
		log.Fatalf("error creating sys-account service: %v", err)
	}

	// AuthRepo will be the object to be passed around other services if you will
	// TODO @gutterbacon: Once config is here, source one from the yamls
	recoveryOptions := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(recoveryHandler(log)),
	}

	logrusOpts := []grpcLogrus.Option{
		grpcLogrus.WithLevels(grpcLogrus.DefaultCodeToLevel),
	}

	unaryItc := []grpc.UnaryServerInterceptor{
		grpcRecovery.UnaryServerInterceptor(recoveryOptions...),
		grpcLogrus.UnaryServerInterceptor(log, logrusOpts...),
	}

	streamItc := []grpc.StreamServerInterceptor{
		grpcRecovery.StreamServerInterceptor(recoveryOptions...),
		grpcLogrus.StreamServerInterceptor(log, logrusOpts...),
	}

	unaryItc, streamItc = svc.InjectInterceptors(unaryItc, streamItc)
	grpcSrv := grpc.NewServer(
		grpcMw.WithUnaryServerChain(unaryItc...),
		grpcMw.WithStreamServerChain(streamItc...),
	)

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
			log.Infof("Request Endpoint: %s", r.URL)
			grpcWebServer.ServeHTTP(w, r)
		}), &http2.Server{}),
	}
	httpServer.Addr = fmt.Sprintf("%s:%d", defaultAddr, defaultPort)
	log.Infof("service listening at %v\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("error running http service: %v\n", err)
	}
}
