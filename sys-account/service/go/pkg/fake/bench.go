// bench is just a benchmark package
package fake

import (
	sharePkg "github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	benchPkg "github.com/getcouragenow/sys-share/sys-core/service/bench"
	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	defaultJSONOutputPath = "./bench/fake-register-data.json"
	defaultHost           = "127.0.0.1:8888"
	defaultProtoPath      = "../sys-share/sys-account/proto/v2/sys_account_services.proto"
	defaultSvcName        = "v2.sys_account.services.AuthService.Register"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sys-bench",
		Short: "running sys-account bench",
	}
	jsonOutPath string
	hostAddr    string
	protoPath   string
	svcName     string
)

func SysAccountBench() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&jsonOutPath, "json-out-path", "j", defaultJSONOutputPath, "default output path for generated fake json data")
	rootCmd.PersistentFlags().StringVarP(&hostAddr, "host-address", "s", defaultHost, "host address to connect: ex 127.0.0.1:8888")
	rootCmd.PersistentFlags().StringVarP(&protoPath, "protobuf-service-path", "p", defaultProtoPath, "protobuf service definition path in filesystem")
	rootCmd.PersistentFlags().StringVarP(&svcName, "service-name", "n", defaultSvcName, "service name to call: ex: "+defaultSvcName)

	l := logrus.New().WithField("svc", "sys-bench")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		fakeRegistersReqs := sharePkg.NewFakeRegisterRequests()
		if err := fakeRegistersReqs.ToJSON(jsonOutPath); err != nil {
			l.Errorf("error generating fake json data: %v", err)
			return err
		}
		dirPath := filepath.Dir(jsonOutPath)
		exists, _ := sharedConfig.PathExists(dirPath)
		if !exists {
			os.MkdirAll(dirPath, 0755)
		}
		if err := benchPkg.RunBench(
			svcName,
			defaultHost,
			jsonOutPath,
			protoPath,
			1000,
			100,
		); err != nil {
			l.Errorf("error running benchmark data: %v", err)
			return err
		}
		return nil
	}

	return rootCmd
}
