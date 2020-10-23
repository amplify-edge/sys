package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	sharedpkg "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
)

var rootCmd = &cobra.Command{
	Use:   "core",
	Short: "core cli",
}

func main() {
	rootCmd.AddCommand(
		sharedpkg.NewSysCoreProxyClient().CobraCommand(),
		sharedpkg.NewSysBusProxyClient().CobraCommand(),
	)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
