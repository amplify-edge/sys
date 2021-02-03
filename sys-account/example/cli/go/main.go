package main

import (
	"log"

	"github.com/amplify-cms/sys-share/sys-account/service/go/pkg"
)

func main() {
	spsc := pkg.NewSysShareProxyClient()
	rootCmd := spsc.CobraCommand()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}

	// Extend it here for local thing.
	// TODO @gutterbacon: do this once Config is here.
}
