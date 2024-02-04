package cmd

import (
	"github.com/spf13/cobra"
)

func NewVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Set up fsverify",
		RunE:         SetupCommand,
		SilenceUsage: true,
	}

	return cmd
}

func SetupCommand(_ *cobra.Command, args []string) error {
	return nil
}
