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

	/*	node := core.Node{
			BlockStart:  0,
			BlockEnd:    4000,
			BlockSum:    "ba0064e29f79feddc3b7912c697a80c93ada98a916b19573ff41598c17177b92",
			PrevNodeSum: "Entrypoint",
		}

		err := core.AddNode(node, nil)
		if err != nil {
			return err
		}*/

	header, err := core.ReadHeader("/dev/sda")
	fmt.Printf("Magic Number: %d\n", header.MagicNumber)
	fmt.Printf("Signature: %s", header.Signature)
	fmt.Printf("FsSize: %d\n", header.FilesystemSize)
	fmt.Printf("FsUnit: %d\n", header.FilesystemUnit)
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

	getnode, err := core.GetNode("Entrypoint", db)
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
	if err != nil {
		return err
	}

	err = core.VerifyBlock(part, getnode)
	if err != nil {
		fmt.Println("fail")
		return err
	}
	fmt.Printf("Block '%s' ranging from %d to %d matches!\n", getnode.PrevNodeSum, getnode.BlockStart, getnode.BlockEnd)

	fmt.Println("----")

	key, err := core.ReadKey()
	if err != nil {
		return err
	}
	fmt.Println("Key: " + key)

	err = core.VerifySignature(key, header.Signature, dbfile)
	if err != nil {
		return err
	} else {
		fmt.Println("Signtaure success")
	}
	return nil
}
