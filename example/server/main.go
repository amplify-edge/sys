package main

import (
	"fmt"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/certutils"
	corebus "go.amplifyedge.org/sys-share-v2/sys-core/service/go/pkg/bus"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging/zaplog"
	grpcMw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"go.amplifyedge.org/sys-v2/main/pkg"
)

const (
	errSourcingConfig   = "error while sourcing config for %s: %v"
	errCreateSysService = "error while creating sys-* service: %v"

	defaultPort                 = 8888
	defaultSysAccountConfigPath = "./config/sysaccount.yml"
	defaultLocalTLSCert         = "./certs/local.pem"
	defaultLocalTLSKey          = "./certs/local.key.pem"
	defaultTLSEnabled           = true
)

var (
	rootCmd          = &cobra.Command{Use: "sys-ex-server"}
	accountCfgPath   string
	mainexPort       int
	tlsEnabled       bool
	localTlsCertPath string
	localTlsKeyPath  string
)

func main() {
	// persistent flags
	rootCmd.PersistentFlags().StringVarP(&accountCfgPath, "sys-account-config-path", "a", defaultSysAccountConfigPath, "sys-account config path to use")
	rootCmd.PersistentFlags().StringVarP(&localTlsCertPath, "tls-cert-path", "t", defaultLocalTLSCert, "local TLS Cert path")
	rootCmd.PersistentFlags().StringVarP(&localTlsKeyPath, "tls-key-path", "k", defaultLocalTLSKey, "local TLS Key path")
	rootCmd.PersistentFlags().IntVarP(&mainexPort, "port", "p", defaultPort, "grpc port to run")
	rootCmd.PersistentFlags().BoolVarP(&tlsEnabled, "enable-tls", "s", defaultTLSEnabled, "enable TLS")

	// logging
	logger := zaplog.NewZapLogger(zaplog.DEBUG, "sys-all", true, "")
	logger.InitLogger(nil)

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// configs
		sspaths := pkg.NewServiceConfigPaths(accountCfgPath, nil)
		cbus := corebus.NewCoreBus()
		sscfg, err := pkg.NewSysServiceConfig(logger, nil, sspaths, defaultPort, cbus)
		if err != nil {
			logger.Fatalf(errSourcingConfig, err)
		}

		// initiate all sys-* service
		sysSvc, err := pkg.NewService(sscfg, "127.0.0.1")
		if err != nil {
			logger.Fatalf(errCreateSysService, err)
		}

		// initiate grpc server
		unaryInterceptors, streamInterceptors := sysSvc.InjectInterceptors(nil, nil)
		var grpcServer *grpc.Server
		if tlsEnabled {
			logger.Info("Server Running With TLS Enabled")
			tlsCreds, err := certutils.LoadTLSKeypair(localTlsCertPath, localTlsKeyPath)
			if err != nil {
				logger.Fatalf(errCreateSysService, err)
			}
			grpcServer = grpc.NewServer(
				grpc.Creds(tlsCreds),
				grpcMw.WithUnaryServerChain(unaryInterceptors...),
				grpcMw.WithStreamServerChain(streamInterceptors...),
			)
		} else {
			logger.Info("Server Running With TLS Disabled")
			grpcServer = grpc.NewServer(
				grpcMw.WithUnaryServerChain(unaryInterceptors...),
				grpcMw.WithStreamServerChain(streamInterceptors...),
			)
		}

		sysSvc.RegisterServices(grpcServer)
		grpcWebServer := sysSvc.RegisterGrpcWebServer(grpcServer)
		// run server
		return sysSvc.Run(fmt.Sprintf("%s:%d", "127.0.0.1", mainexPort), grpcWebServer, nil, localTlsCertPath, localTlsKeyPath)
	}
	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error running sys-main: %v", err)
	}
}
