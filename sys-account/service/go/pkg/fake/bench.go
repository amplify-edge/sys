// bench is just a benchmark package
package fake

import (
	"fmt"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/fileutils"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging/zaplog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	sharePkg "go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg"
	benchPkg "go.amplifyedge.org/sys-share-v2/sys-core/service/bench"
)

const (
	defaultJSONOutputPath  = "./bench/fake-register-data.json"
	defaultHost            = "127.0.0.1:8888"
	defaultProtoPath       = "../sys-share/sys-account/proto/v2/sys_account_services.proto"
	defaultSvcName         = "v2.sys_account.services.AuthService.Register"
	defaultNumberOfRecords = 100
	defaultConcurrency     = 10
	defaultTlsEnabled      = false
	defaultTlsCertPath     = "./rootca.pem"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sys-bench",
		Short: "running sys-account bench",
	}
	jsonOutPath       string
	hostAddr          string
	protoPath         string
	svcName           string
	recordNumber      uint
	concurrentRequest uint
	tlsEnabled        bool
	tlsCaCertPath     string
)

func SysAccountBench() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&jsonOutPath, "json-out-path", "j", defaultJSONOutputPath, "default output path for generated fake json data")
	rootCmd.PersistentFlags().StringVarP(&hostAddr, "host-address", "s", defaultHost, "host address to connect: ex 127.0.0.1:8888")
	rootCmd.PersistentFlags().StringVarP(&protoPath, "protobuf-service-path", "p", defaultProtoPath, "protobuf service definition path in filesystem")
	rootCmd.PersistentFlags().StringVarP(&svcName, "service-name", "n", defaultSvcName, "service name to call: ex: "+defaultSvcName)
	rootCmd.PersistentFlags().UintVarP(&recordNumber, "number-of-records", "r", defaultNumberOfRecords, fmt.Sprintf("number of records to test, default: %d", defaultNumberOfRecords))
	rootCmd.PersistentFlags().UintVarP(&concurrentRequest, "concurrent-requests", "c", defaultConcurrency, fmt.Sprintf("number of concurrent requests, default: %d", defaultConcurrency))
	rootCmd.PersistentFlags().BoolVarP(&tlsEnabled, "enable-tls", "e", defaultTlsEnabled, "enable tls")
	rootCmd.PersistentFlags().StringVarP(&tlsCaCertPath, "tls-cert-path", "t", defaultTlsCertPath, "CA Cert Path")

	l := zaplog.NewZapLogger(zaplog.DEBUG, "sys-bench", true, "")
	l.InitLogger(nil)

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		fakeRegistersReqs := sharePkg.NewFakeRegisterRequests()
		if err := fakeRegistersReqs.ToJSON(jsonOutPath); err != nil {
			l.Errorf("error generating fake json data: %v", err)
			return err
		}
		dirPath := filepath.Dir(jsonOutPath)
		exists, _ := fileutils.PathExists(dirPath)
		if !exists {
			os.MkdirAll(dirPath, 0755)
		}
		if err := benchPkg.RunBench(
			svcName,
			defaultHost,
			jsonOutPath,
			protoPath,
			recordNumber,
			concurrentRequest,
			tlsEnabled,
			tlsCaCertPath,
		); err != nil {
			l.Errorf("error running benchmark data: %v", err)
			return err
		}
		return nil
	}

	return rootCmd
}
