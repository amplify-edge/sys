// package main
// this is only used for local testing
// making sure that the sys-account server works locally before wiring it up to the maintemplate.
package main

import (
	"context"
	"github.com/getcouragenow/sys/sys-account/server"
	"github.com/getcouragenow/sys/sys-core/server/pkg/db"

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
	"google.golang.org/grpc/reflection"

	rpc "github.com/getcouragenow/sys-share/sys-account/server/rpc/v2"

	"github.com/getcouragenow/sys/sys-account/server/delivery"
	"github.com/getcouragenow/sys/sys-account/server/pkg/utilities"
)

func recoveryHandler(l *logrus.Entry) func(panic interface{}) error {
	return func(panic interface{}) error {
		l.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

var (
	defaultUnauthenticatedRoutes = []string{"/getcouragenow.sys.v2.sys_account.AuthService/Login",
		"/getcouragenow.sys.v2.sys_account.AuthService/Register",
		"/getcouragenow.sys.v2.sys_account.AuthService/ResetPassword",
		"/getcouragenow.sys.v2.sys_account.AuthService/ForgotPassword",
		"/getcouragenow.sys.v2.sys_account.AuthService/RefreshAccessToken",
		// debugging purposes
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}
)

func main() {
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
	sysAccCfg := &server.SysAccountConfig{
		UnauthenticatedRoutes: defaultUnauthenticatedRoutes,
		JWTConfig: server.JWTConfig{
			Access: server.TokenConfig{
				Secret: string(accessSecret),
			},
			Refresh: server.TokenConfig{
				Secret: string(refreshSecret),
			},
		},
	}
	authDelivery, err := delivery.NewAuthDeli(log, db.SharedDatabase(), sysAccCfg)
	if err != nil {
		log.Fatalf("cannot create auth delivery: %v", err)
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
	rpc.RegisterAuthServiceService(grpcSrv, &rpc.AuthServiceService{
		Register:           authDelivery.Register,
		Login:              authDelivery.Login,
		ForgotPassword:     authDelivery.ForgotPassword,
		ResetPassword:      authDelivery.ResetPasssword,
		RefreshAccessToken: authDelivery.RefreshAccessToken,
	})
	rpc.RegisterAccountServiceService(grpcSrv, &rpc.AccountServiceService{
		GetAccount: authDelivery.GetAccount,
	})
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
			log.Infof("Request Endpoint: %s", r.URL)
			grpcWebServer.ServeHTTP(w, r)
		}), &http2.Server{}),
	}
	httpServer.Addr = "127.0.0.1:8888"
	log.Infof("server listening at %v\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("error running http server: %v\n", err)
	}
}
