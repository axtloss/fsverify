package cmd

import (
	"bytes"
	"fmt"
	verify "github.com/axtloss/fsverify/core"
	"github.com/axtloss/fsverify/verifysetup/core"
	"github.com/spf13/cobra"
	"math"
	"os"
)

func NewSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Set up fsverify",
		RunE:         SetupCommand,
		SilenceUsage: true,
	}

	return cmd
}

func SetupCommand(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Usage: verifysetup setup [partition]")
	}
	fmt.Println("Using partition: ", args[0])
	disk, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer disk.Close()
	fmt.Println("Reading from disk")
	diskInfo, err := disk.Stat()
	if err != nil {
		return err
	}
	diskSize := diskInfo.Size()
	blockCount := math.Floor(float64(diskSize / 4000))
	lastBlockSize := float64(diskSize) - blockCount*4000.0
	fmt.Println(diskSize)
	fmt.Println(blockCount)
	fmt.Println(lastBlockSize)
	node := verify.Node{}
	block := make([]byte, 4000)
	diskBytes := make([]byte, diskSize)
	_, err = disk.Read(diskBytes)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(diskBytes)
	for i := 0; i < int(blockCount); i++ {
		reader = bytes.NewReader(diskBytes)
		block, err = core.ReadBlock(i*4000, (i*4000)+4000, reader)
		if err != nil {
			return err
		}
		node, err = core.CreateNode(i*4000, (i*4000)+4000, block, &node)
		if err != nil {
			return err
		}
		fmt.Println(node)
		err = core.AddNode(node, nil, "./fsverify.db")
	}
	finalBlock, err := core.ReadBlock(int(blockCount*4000), int((blockCount*4000)+lastBlockSize), reader)
	if err != nil {
		return err
	}
	finalNode, err := core.CreateNode(int(blockCount*4000), int((blockCount*4000)+lastBlockSize), finalBlock, &node)
	if err != nil {
		return err
	}
	fmt.Println(finalNode)
	err = core.AddNode(finalNode, nil, "./fsverify.db")
	if err != nil {
		return err
	}

	signature, err := core.SignDatabase("./fsverify.db", "./minisign/")
	if err != nil {
		return err
	}
	fmt.Println(string(signature))
	return nil
}
