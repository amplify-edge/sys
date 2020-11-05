package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	corepkg "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/fake"
)

func main() {
	logger := log.New().WithField("sys-sdk", "sys-*")
	logger.Println(" -- sdk cli -- ")

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
		log.Fatalf("command failed: %v", err)
	}
}
