package main

import (
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging/zaplog"

	"go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg"
	corepkg "go.amplifyedge.org/sys-share-v2/sys-core/service/go/pkg"
	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/fake"
)

func main() {
	zlog := zaplog.NewZapLogger(zaplog.DEBUG, "exampleSys", true, "")
	zlog.InitLogger(nil)
	zlog.Info("starting sys example")

	spsc := pkg.NewSysShareProxyClient()
	rootCmd := spsc.CobraCommand()

	// load up sdk cli
	coreProxyCli := corepkg.NewSysCoreProxyClient()
	busProxyCli := corepkg.NewSysBusProxyClient()
	mailProxyCli := corepkg.NewSysMailProxyClient()
	fileProxyCli := corepkg.NewFileServiceClientCommand()
	rootCmd.AddCommand(fake.SysAccountBench(), coreProxyCli.CobraCommand(), busProxyCli.CobraCommand(), mailProxyCli.CobraCommand(), fileProxyCli)

	// starts proxy
	if err := rootCmd.Execute(); err != nil {
		zlog.Fatalf("command failed: %v", err)
	}
}
