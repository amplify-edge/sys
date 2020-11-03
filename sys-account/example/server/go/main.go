// package main
// this is only used for local testing
// making sure that the sys-account service works locally before wiring it up to the maintemplate.
package main

import (
	"fmt"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/mailer"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	grpcMw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	corebus "github.com/getcouragenow/sys-share/sys-core/service/go/pkg/bus"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg"
	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

const (
	errSourcingConfig   = "error while sourcing config for %s: %v"
	errCreateSysService = "error while creating sys-%s service: %v"

	defaultAddr                 = "127.0.0.1"
	defaultPort                 = 8888
	defaultSysCoreConfigPath    = "./bin-all/config/syscore.yml"
	defaultSysAccountConfigPath = "./bin-all/config/sysaccount.yml"
)

var (
	rootCmd        = &cobra.Command{Use: "sys-account-srv"}
	coreCfgPath    string
	accountCfgPath string
)

func recoveryHandler(l *logrus.Entry) func(panic interface{}) error {
	return func(panic interface{}) error {
		l.Warnf("sys-account service recovered, reason: %v",
			panic)
		return nil
	}
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&coreCfgPath, "sys-core-config-path", "c", defaultSysCoreConfigPath, "sys-core config path to use")
	rootCmd.PersistentFlags().StringVarP(&accountCfgPath, "sys-account-config-path", "a", defaultSysAccountConfigPath, "sys-account config path to use")

	log := logrus.New().WithField("svc", "sys-account")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		csc, err := corecfg.NewConfig(coreCfgPath)
		if err != nil {
			log.Fatalf(errSourcingConfig, err)
		}

		gdb, err := coredb.NewCoreDB(log, &csc.SysCoreConfig, nil)
		if err != nil {
			log.Fatalf(errSourcingConfig, err)
		}

		newMailSvc := mailer.NewMailSvc(&csc.MailConfig, log)

		sysAccountConfig, err := accountpkg.NewSysAccountServiceConfig(log, gdb, accountCfgPath, corebus.NewCoreBus(), newMailSvc)
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}

		svc, err := accountpkg.NewSysAccountService(sysAccountConfig)
		if err != nil {
			log.Fatalf("error creating sys-account service: %v", err)
		}

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

		// Register sys-account service
		svc.RegisterServices(grpcSrv)

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
		return httpServer.ListenAndServe()
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
