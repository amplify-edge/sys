package main

import (
	"github.com/getcouragenow/sys/main/pkg"
	"github.com/sirupsen/logrus"
)

var (
	defaultUnauthenticatedRoutes = []string{
		"/v2.services.AuthService/Login",
		"/v2.services.AuthService/Register",
		"/v2.services.AuthService/ResetPassword",
		"/v2.services.AuthService/ForgotPassword",
		"/v2.services.AuthService/RefreshAccessToken",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}
)

const (
	defaultPort         = 8888
	errSourcingConfig   = "error while sourcing config: %v"
	errCreateSysService = "error while creating sys-* service: %v"
)

func main() {
	logger := logrus.New().WithField("sys-main", "sys-*")
	sscfg, err := pkg.NewSysServiceConfig(nil, defaultUnauthenticatedRoutes, defaultPort)
	if err != nil {
		logger.Fatalf(errSourcingConfig, err)
	}
	sysSvc, err := pkg.NewService(sscfg)
	if err != nil {
		logger.Fatalf(errCreateSysService, err)
	}
	sysSvc.Run(nil, nil)
}
