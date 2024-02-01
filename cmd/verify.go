package cmd

import (
	"fmt"
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

	header, err := core.ReadHeader("./part.fsverify")
	fmt.Println(header.MagicNumber)
	fmt.Println(header.Signature)
	fmt.Println(header.FilesystemSize)
	fmt.Println(header.TableSize)
	if err != nil {
		return err
	}
	dbfile, err := core.ReadDB("./part.fsverify")
	if err != nil {
		return err
	}
	fmt.Println("DBFILE: ", dbfile)
	db, err := core.OpenDB(dbfile)
	if err != nil {
		return err
	}

	getnode, err := core.GetNode("aaaa", db)
	if err != nil {
		return err
	}
	fmt.Println(getnode)
	return nil
}
