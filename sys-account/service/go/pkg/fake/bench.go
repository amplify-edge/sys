// bench is just a benchmark package
package fake

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	sharePkg "github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	benchPkg "github.com/getcouragenow/sys-share/sys-core/service/bench"
)

const (
	defaultJSONOutputPath = "./bench/fake-register-data.json"
	defaultHost           = "127.0.0.1:8888"
	defaultProtoPath      = "../sys-share/sys-account/proto/v2/services.proto"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sys-bench",
		Short: "running sys-account bench",
	}
	jsonOutPath string
	hostAddr    string
	protoPath   string
)

func SysAccountBench() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&jsonOutPath, "json-out-path", "j", defaultJSONOutputPath, "default output path for generated fake json data")
	rootCmd.PersistentFlags().StringVarP(&hostAddr, "host-address", "s", defaultHost, "host address to connect: ex 127.0.0.1:8888")
	rootCmd.PersistentFlags().StringVarP(&protoPath, "protobuf-service-path", "p", defaultProtoPath, "protobuf service definition path in filesystem")

	l := logrus.New().WithField("svc", "sys-bench")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		fakeRegistersReqs := sharePkg.NewFakeRegisterRequests()
		if err := fakeRegistersReqs.ToJSON(jsonOutPath); err != nil {
			l.Errorf("error generating fake json data: %v", err)
			return err
		}
		if err := benchPkg.RunBench(
			"v2.services.AuthService.Register",
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
