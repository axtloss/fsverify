package cmd

import (
	"bufio"
	"fmt"
	"github.com/axtloss/fsverify/core"
	"github.com/spf13/cobra"
	"os"
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

	header, err := core.ReadHeader("/dev/sda")
	fmt.Printf("Magic Number: %d\n", header.MagicNumber)
	fmt.Printf("Signature: %s\n" + header.Signature)
	fmt.Printf("FsSize: %d\n", header.FilesystemSize)
	fmt.Printf("Table Size: %d\n", header.TableSize)
	fmt.Printf("Table Size Unit: %d\n", header.TableUnit)
	if err != nil {
		return err
	}
	dbfile, err := core.ReadDB("/dev/sda")
	if err != nil {
		return err
	}
	fmt.Println("DBFILE: ", dbfile)
	db, err := core.OpenDB(dbfile, true)
	if err != nil {
		return err
	}

	getnode, err := core.GetNode("aaaa", db)
	if err != nil {
		return err
	}
	fmt.Println(getnode)

	fmt.Println("----")

	disk, err := os.Open("./partition.raw")
	reader := bufio.NewReader(disk)
	part, err := core.ReadBlock(getnode, reader)
	if err != nil {
		return err
	}
	hash, err := core.CalculateBlockHash(part)
	fmt.Println(hash)
	return err
}
