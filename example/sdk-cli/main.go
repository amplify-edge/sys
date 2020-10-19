package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/fake"
)

func main() {
	log.Println(" -- sdk cli -- ")
	// load up sdk cli
	spsc := pkg.NewSysShareProxyClient()
	rootCmd := spsc.CobraCommand()
	rootCmd.AddCommand(fake.SysAccountBench())

	// starts proxy
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
