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

	/*entrynode := core.Node{
		BlockStart:  0,
		BlockEnd:    4000,
		BlockSum:    "32fd1c42b66cbf1b2f0f1a65a3cb08f3d7845eac7f43e13b2b5b5f9f837e3346",
		PrevNodeSum: "Entrypoint",
	}

	err := core.AddNode(entrynode, nil)
	if err != nil {
		return err
	}

	entryHash, err := entrynode.GetHash()
	if err != nil {
		return err
	}
	nextNode := core.Node{
		BlockStart:  4000,
		BlockEnd:    8000,
		BlockSum:    "3d73ff8cb154dcfe8cdae426021f679e541b47dbe14e8426e6b1cd3f2c57017c",
		PrevNodeSum: entryHash,
	}

	err = core.AddNode(nextNode, nil)
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

	fmt.Println("----")

	disk, err := os.Open("./partition.raw")
	reader := bufio.NewReader(disk)
	part, err := core.ReadBlock(getnode, reader)
	if err != nil {
		return err
	}
	diskInfo, err := disk.Stat()
	node, err := core.GetNode("Entrypoint", db)
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

	fmt.Println(node)
	for int64(core.TotalReadBlocks) < diskInfo.Size() {
		nodeSum, err := node.GetHash()
		if err != nil {
			return err
		}
		node, err := core.GetNode(nodeSum, db)
		if err != nil {
			return err
		}
		fmt.Println("----")
		fmt.Println(node)
		part, err := core.ReadBlock(node, reader)
		if err != nil {
			return err
		}
		hash, err := core.CalculateBlockHash(part)
		fmt.Println(hash)
		if err != nil {
			return err
		}
		err = core.VerifyBlock(part, node)
		if err != nil {
			fmt.Println("fail")
			return err
		}
		fmt.Printf("Block '%s' ranging from %d to %d matches!\n", node.PrevNodeSum, node.BlockStart, node.BlockEnd)

	}

	return nil
}
