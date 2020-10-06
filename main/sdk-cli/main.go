package main

import (
	"github.com/getcouragenow/sys-share/pkg"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println(" -- sdk cli -- ")
	// load up sdk cli
	spsc := pkg.NewSysShareProxyClient()
	rootCmd := spsc.CobraCommand()

	// starts proxy
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
