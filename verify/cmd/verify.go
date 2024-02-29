package cmd

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/axtloss/fsverify/verify/config"
	"github.com/axtloss/fsverify/verify/core"
	"github.com/spf13/cobra"
)

var validateFailed bool

func NewVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "verify",
		Short:        "Verify the root filesystem based on the given verification",
		RunE:         ValidateCommand,
		SilenceUsage: true,
	}

	return cmd
}

// validateThread validates a chain of nodes against a given byte slice
func validateThread(blockStart int, blockEnd int, bundleSize int, diskBytes []byte, n int, dbfile string, waitGroup *sync.WaitGroup, errChan chan error) {
	defer waitGroup.Done()
	defer close(errChan)
	var reader *bytes.Reader
	blockCount := math.Floor(float64(bundleSize / 2000))
	totalReadBlocks := 0

	db, err := core.OpenDB(dbfile, true)
	if err != nil {
		errChan <- err
	}

	reader = bytes.NewReader(diskBytes)

	node, err := core.GetNode(fmt.Sprintf("Entrypoint%d", n), db)
	if err != nil {
		errChan <- err
	}
	block, i, err := core.ReadBlock(node, reader, totalReadBlocks)
	totalReadBlocks = i

	err = core.VerifyBlock(block, node)
	if err != nil {
		errChan <- err
	}

	for int64(totalReadBlocks) < int64(blockCount) {
		if validateFailed {
			return
		}
		nodeSum, err := node.GetHash()
		if err != nil {
			fmt.Println("Using node ", nodeSum)
			errChan <- err
		}
		node, err = core.GetNode(nodeSum, db)
		if err != nil {
			fmt.Println("Failed to get next node")
			errChan <- err
		}
		part, i, err := core.ReadBlock(node, reader, totalReadBlocks)
		totalReadBlocks = i
		if err != nil {
			errChan <- err
			validateFailed = true
			return
		}
		err = core.VerifyBlock(part, node)
		if err != nil {
			errChan <- err
			validateFailed = true
			return
		}

	}

}

func ValidateCommand(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Usage: fsverify verify [disk]")
	}

	header, err := core.ReadHeader(config.FsVerifyPart)
	if err != nil {
		return err
	}

	// Check if the partition is even correct
	// this does not check if the partition has been tampered with
	// it only checks if the specified partition is even an fsverify partition
	if header.MagicNumber != 0xACAB {
		return fmt.Errorf("sanity bit does not match. Expected %d, got %d", 0xACAB, header.MagicNumber)
	}

	fmt.Println("Reading DB")
	dbfile, err := core.ReadDB(config.FsVerifyPart)
	if err != nil {
		return err
	}
	key, err := core.ReadKey()
	if err != nil {
		return err
	}
	fmt.Println("Key: " + key)
	verified, err := core.VerifySignature(key, header.Signature, dbfile)
	if err != nil {
		return err
	} else if !verified {
		return fmt.Errorf("Signature verification failed\n")
	} else {
		fmt.Println("Signature verification success!")
	}

	fmt.Println("----")
	disk, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer disk.Close()
	diskInfo, err := disk.Stat()
	if err != nil {
		return err
	}
	diskSize := diskInfo.Size()

	// If the filesystem size has increased ever since the fsverify partition was created
	// it would mean that fsverify is not able to verify the entire partition, making it useless
	if header.FilesystemSize*header.FilesystemUnit != int(diskSize) {
		return fmt.Errorf("disk size does not match disk size specified in header. Expected %d, got %d", header.FilesystemSize*header.FilesystemUnit, diskSize)
	}

	bundleSize := math.Floor(float64(diskSize / int64(config.ProcCount)))
	diskBytes := make([]byte, diskSize)
	_, err = disk.Read(diskBytes)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(diskBytes)
	var waitGroup sync.WaitGroup
	errChan := make(chan error)
	validateFailed = false
	for i := 0; i < config.ProcCount; i++ {
		// To ensure that each thread only uses the byte area it is meant to use, a copy of the
		// area is made
		diskBytes, err := core.CopyByteArea(i*(int(bundleSize)), (i+1)*(int(bundleSize)), reader)
		if err != nil {
			fmt.Println("Failed to copy byte area ", i*int(bundleSize), " ", (i+1)+int(bundleSize))
			return err
		}
		waitGroup.Add(1)
		go validateThread(i*int(bundleSize), (i+1)*int(bundleSize), int(bundleSize), diskBytes, i, dbfile, &waitGroup, errChan)
	}

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			core.WarnUser()
			return err
		}
	}

	return nil
}
