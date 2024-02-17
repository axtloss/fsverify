package cmd

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/axtloss/fsverify/config"
	"github.com/axtloss/fsverify/core"
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

func validateThread(blockStart int, blockEnd int, bundleSize int, diskBytes []byte, n int, dbfile string, waitGroup *sync.WaitGroup, errChan chan error) {
	defer waitGroup.Done()
	defer close(errChan)
	var reader *bytes.Reader
	blockCount := math.Floor(float64(bundleSize / 2000))
	totalReadBlocks := 0

	fmt.Println("DBFILE: ", dbfile)
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
		fmt.Println("fail")
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
		node, err := core.GetNode(nodeSum, db)
		if err != nil {
			errChan <- err
		}
		fmt.Println("----")
		fmt.Println(node)
		part, i, err := core.ReadBlock(node, reader, totalReadBlocks)
		totalReadBlocks = i
		if err != nil {
			errChan <- err
			validateFailed = true
			return
		}
		err = core.VerifyBlock(part, node)
		if err != nil {
			fmt.Println("fail")
			errChan <- err
			validateFailed = true
			return
			//fmt.Printf("Block '%s' ranging from %d to %d matches!\n", node.PrevNodeSum, node.BlockStart, node.BlockEnd)
		}

	}

}

func ValidateCommand(_ *cobra.Command, args []string) error {
	header, err := core.ReadHeader("./part.fsverify")
	fmt.Printf("Magic Number: %d\n", header.MagicNumber)
	fmt.Printf("Signature: %s", header.Signature)
	fmt.Printf("FsSize: %d\n", header.FilesystemSize)
	fmt.Printf("FsUnit: %d\n", header.FilesystemUnit)
	fmt.Printf("Table Size: %d\n", header.TableSize)
	fmt.Printf("Table Size Unit: %d\n", header.TableUnit)
	if err != nil {
		return err
	}
	fmt.Println("Reading DB")
	//dbfile, err := core.ReadDB("/dev/sda")
	dbfile, err := core.ReadDB("./part.fsverify")
	if err != nil {
		return err
	}
	fmt.Println("DBFILE: ", dbfile)
	/*	db, err := core.OpenDB(dbfile, true)
		if err != nil {
			return err
		}*/

	key, err := core.ReadKey()
	if err != nil {
		return err
	}
	fmt.Println("Key: " + key)
	verified, err := core.VerifySignature(key, header.Signature, dbfile)
	if err != nil {
		return err
	} else if !verified {
		//return fmt.Errorf("Signature verification failed\n")
		fmt.Println("Signature verification failedw")
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

	bundleSize := math.Floor(float64(diskSize / int64(config.ProcCount)))
	//	blockCount := math.Ceil(float64(bundleSize / 2000))
	//	lastBlockSize := int(diskSize) - int(diskSize)*config.ProcCount
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
			return err
		}
	}

	/*for int64(core.TotalReadBlocks) < diskInfo.Size() {
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

	}*/

	return nil
}
