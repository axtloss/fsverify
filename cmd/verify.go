package cmd

import (
	"github.com/axtloss/fsverify/core"
	"github.com/spf13/cobra"
)

func NewVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "verify",
		Short:        "Verify the root filesystem based on the given verification",
		RunE:         ValidateCommand,
		SilenceUsage: true,
	}

	return cmd
}

func ValidateCommand(_ *cobra.Command, args []string) error {
	node := core.Node{
		BlockStart:  0,
		BlockEnd:    4 * 1000,
		BlockSum:    "test",
		PrevNodeSum: "aaaa",
	}
	err := core.AddNode(node, nil)
	if err != nil {
		return err
	}
	_, err = core.ReadHeader("./test.part")
	return err
}
