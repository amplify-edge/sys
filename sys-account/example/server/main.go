// package main
// this is only used for local testing
// making sure that the sys-account service works locally before wiring it up to the maintemplate.
package main

import (
	"context"
	"github.com/genjidb/genji"
	server "github.com/getcouragenow/sys/sys-account/service/go"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
	"net/http"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/getcouragenow/sys-share/pkg"

	"github.com/getcouragenow/sys/sys-account/service/go/delivery"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/utilities"
)

var (
	gdb *genji.DB
)

func recoveryHandler(l *logrus.Entry) func(panic interface{}) error {
	return func(panic interface{}) error {
		l.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

func initdb() {
	gdb = coredb.SharedDatabase()
}

func main() {
	initdb()
	log := logrus.New().WithField("svc", "sys-account")
	accessSecret, err := utilities.GenRandomByteSlice(32)
	if err != nil {
		log.Fatalf("error creating jwt access token secret: %v\n", err)
		os.Exit(1)
	}
	refreshSecret, err := utilities.GenRandomByteSlice(32)
	if err != nil {
		log.Fatalf("error creating jwt access token secret: %v\n", err)
		os.Exit(1)
	}

	// AuthDelivery will be the object to be passed around other services if you will
	// TODO @gutterbacon: Once config is here, source one from the yamls
	accCfg := &server.SysAccountConfig{
		UnauthenticatedRoutes: []string{
			"/v2.services.AuthService/Login",
			"/v2.services.AuthService/Register",
			"/v2.services.AuthService/ResetPassword",
			"/v2.services.AuthService/ForgotPassword",
			"/v2.services.AuthService/RefreshAccessToken",
			// debugging purposes
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
		JWTConfig: server.JWTConfig{
			Access:  server.TokenConfig{Secret: string(accessSecret)},
			Refresh: server.TokenConfig{Secret: string(refreshSecret)},
		},
	}
	authDelivery, err := delivery.NewAuthDeli(log, gdb, accCfg)
	if err != nil {
		log.Fatal(err)
	}

	recoveryOptions := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryHandler(log)),
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	grpcSrv := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(recoveryOptions...),
			grpc_logrus.UnaryServerInterceptor(log, logrusOpts...),
			grpc_auth.UnaryServerInterceptor(authDelivery.DefaultInterceptor),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(recoveryOptions...),
			grpc_logrus.StreamServerInterceptor(log, logrusOpts...),
			grpc_auth.StreamServerInterceptor(authDelivery.DefaultInterceptor),
		),
	)
	sysAccProxy := pkg.NewSysShareProxyService(authDelivery, authDelivery)
	sysAccProxy.RegisterSvc(grpcSrv)

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
	httpServer.Addr = "127.0.0.1:8888"
	log.Infof("service listening at %v\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("error running http service: %v\n", err)
	}
}
