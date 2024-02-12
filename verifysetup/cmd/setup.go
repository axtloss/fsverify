package cmd

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"

	verify "github.com/axtloss/fsverify/core"
	"github.com/axtloss/fsverify/verifysetup/core"
	"github.com/spf13/cobra"
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

func checksumBlock(blockStart int, blockEnd int, blockCount int, diskBytes []byte, nodeChannel chan verify.Node, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var reader *bytes.Reader
	node := verify.Node{}
	//fmt.Printf("Starting from %d to %d. BlockCount is %d\n", blockStart, blockEnd, blockCount)
	//fmt.Println(blockCount)
	//fmt.Println("diskBytes: ")
	//fmt.Printf("Addres of diskBytes: %d\n", &diskBytes)
	//fmt.Printf("%d:: diskByteslen: %d\n", blockStart, len(diskBytes))
	for i := 0; i < int(blockCount)-1; i++ {
		reader = bytes.NewReader(diskBytes)
		block, err := core.ReadBlock(i*2000, (i*2000)+2000, reader)
		if err != nil {
			fmt.Printf("%d:: %d attempted reading from %d to %d. Error %s\n", blockStart, i, i*2000, (i*2000)+2000, err)
			//fmt.Println(err)
			return
		}
		node, err = core.CreateNode(i*2000, (i*2000)+2000, block, &node)
		if err != nil {
			fmt.Printf("%d:: 2 Error %s\n", blockStart, err)
			//fmt.Println(err)
			return
		}
		//nodeChannel <- node
		//fmt.Println(blockStart, ":: ", node)
		//fmt.Printf("%d:: %d\n", blockStart, i)
	}
	fmt.Printf("Node from %d to %d finished.\n", blockStart, blockEnd)
}

func copyByteArea(start int, end int, reader *bytes.Reader) ([]byte, error) {
	bytes := make([]byte, end-start)
	//reader.Seek(int64(start), 0)
	n, err := reader.ReadAt(bytes, int64(start))
	//fmt.Printf("Reading from %d to %d\n", start, end)
	if err != nil {
		return nil, err
	} else if n != end-start {
		return nil, fmt.Errorf("Unable to read requested size. Got %d, expected %d", n, end-start)
	}
	return bytes, nil
}

func SetupCommand(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: verifysetup setup [partition] [procCount]")
	}
	procCount, err := strconv.Atoi(args[1])
	if err != nil {
		return err
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
	blockCount := math.Floor(float64(diskSize / 2000))
	lastBlockSize := float64(diskSize) - blockCount*2000.0
	blockBundle := math.Floor(float64(blockCount / float64(procCount)))
	//	lastBlockBundle := float64(blockCount) - blockBundle*float64(procCount)
	fmt.Println(diskSize)
	fmt.Println(blockCount)
	fmt.Println(lastBlockSize)
	//	node := verify.Node{}
	//	block := make([]byte, 2000)
	diskBytes := make([]byte, diskSize)
	_, err = disk.Read(diskBytes)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(diskBytes)
	nodeChannel := make(chan verify.Node)
	var waitGroup sync.WaitGroup
	//var nodes []verify.Node
	for i := 0; i < procCount; i++ {
		/*reader = bytes.NewReader(diskBytes)
		block, err = core.ReadBlock(i*2000, (i*2000)+2000, reader)
		if err != nil {
			return err
		}
		node, err = core.CreateNode(i*2000, (i*2000)+2000, block, &node)
		if err != nil {
			return err
		}
		fmt.Println(node)
		err = core.AddNode(node, nil, "./fsverify.db")*/

		diskBytesCopy, err := copyByteArea(i*(int(blockBundle)*2000), (i+1)*(int(blockBundle)*2000), reader)
		if err != nil {
			return err
		}
		waitGroup.Add(1)
		fmt.Printf("Starting thread %d with blockStart %d and blockEnd %d\n", i, i*(int(blockBundle)*2000), (i+1)*(int(blockBundle)*2000))
		go checksumBlock(i*(int(blockBundle)*2000), (i+1)*(int(blockBundle)*2000), int(blockBundle), diskBytesCopy, nodeChannel, &waitGroup)
	}
	//fmt.Println("Appending nodes")
	/*for i := 0; i < procCount; i++ {
		nodes = append(nodes, <-nodeChannel)
	}*/
	waitGroup.Wait()
	fmt.Println("Created nodelist")
	/*finalBlock, err := core.ReadBlock(int(blockCount*2000), int((blockCount*2000)+lastBlockSize), reader)
	if err != nil {
		return err
	}
	finalNode, err := core.CreateNode(int(blockCount*2000), int((blockCount*2000)+lastBlockSize), finalBlock, &node)
	if err != nil {
		return err
	}
	fmt.Println(finalNode)
	err = core.AddNode(finalNode, nil, "./fsverify.db")
	if err != nil {
		return err
	}*/

	signature, err := core.SignDatabase("./fsverify.db", "./minisign/")
	if err != nil {
		return err
	}
	fmt.Println(string(signature))
	return nil
}
