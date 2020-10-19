package main

import (
	"github.com/spf13/cobra"
	// "github.com/getcouragenow/sys-share/sys-account/service/rpc/v2"
)

var rootCmd = &cobra.Command{
	Use:   "auth",
	Short: "auth cli",
}

func main() {
	/*
		rootCmd.AddCommand(rpc.AuthServiceClientCommand())
		if err := rootCmd.Execute(); err != nil {
			log.Fatalf("command failed: %v", err)
		}
	*/

	// Extend it here for local thing.
	// TODO @gutterbacon: do this once Config is here.
}
