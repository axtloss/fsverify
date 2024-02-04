package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "verifysetup",
}

func init() {
	rootCmd.AddCommand(NewSetupCommand())
}

func Execute() {
	// cobra does not exit with a non-zero return code when failing
	// solution from https://github.com/spf13/cobra/issues/221
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
