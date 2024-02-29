package cmd

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"

	"aead.dev/minisign"
	verify "github.com/axtloss/fsverify/verify/core"
	"github.com/axtloss/fsverify/verifysetup/core"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
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

// checksumBlock is a function to create a chain of Nodes to verify an area of a block device.
// It is meant to be run as a goroutine, taking a waitGroup as a parameter.
func checksumBlock(blockStart int, blockEnd int, bundleSize int, diskBytes []byte, nodeChannel chan verify.Node, n int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	defer close(nodeChannel)
	var reader *bytes.Reader
	node := verify.Node{}

	// A block is 2000 bytes big
	blockCount := math.Floor(float64(bundleSize / 2000))

	for i := 0; i < int(blockCount); i++ {
		reader = bytes.NewReader(diskBytes)
		block, err := core.ReadBlock(i*2000, (i*2000)+2000, reader)
		if err != nil {
			fmt.Printf("%d:: %d attempted reading from %d to %d. Error %s\n", blockStart, i, i*2000, (i*2000)+2000, err)
			return
		}
		node, err = core.CreateNode(i*2000, (i*2000)+2000, block, &node, strconv.Itoa(n))
		if err != nil {
			fmt.Printf("%d:: Attempted creating node for range %d - %d. Error %s\n", blockStart, i*2000, (i*2000)+2000, err)
			return
		}
		nodeChannel <- node
	}

	// Since it is unlikely that the bundleSize is perfectly divisible by 2000
	// a final node has to be created that includes the last amount of bytes
	block, err := core.ReadBlock(int(blockCount*2000), len(diskBytes), reader)
	if err != nil {
		fmt.Printf("%d:: final attempted reading from %d to %d. Error %s\n", blockStart, int(blockCount*2000)+2000, len(diskBytes), err)
		return
	}
	finalNode, err := core.CreateNode(blockStart+int(blockCount*2000)+2000, len(diskBytes), block, &node, strconv.Itoa(n))
	nodeChannel <- finalNode
	fmt.Printf("Node from %d to %d finished.\n", blockStart, blockEnd)
}

func SetupCommand(_ *cobra.Command, args []string) error {
	if len(args) != 3 && len(args) != 4 {
		return fmt.Errorf("Usage: verifysetup setup [partition] [procCount] [fsverify partition output] <minisign directory>")
	}

	// The minisign directory argument is optional
	// which is why the existence of the argument is checked
	// before minisignDir is set to a directory
	var minisignDir string
	if len(args) != 4 {
		minisignDir = "./minisign/"
	} else {
		minisignDir = args[3]
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
	bundleSize := math.Floor(float64(diskSize / int64(procCount)))
	blockCount := math.Ceil(float64(bundleSize / 2000))
	diskBytes := make([]byte, diskSize)
	_, err = disk.Read(diskBytes)
	if err != nil {
		return err
	}

	// To decrease the amount of file operations
	// a single reader is created that gets used for the goroutines
	reader := bytes.NewReader(diskBytes)
	var waitGroup sync.WaitGroup
	nodeChannels := make([]chan verify.Node, procCount+1)
	for i := 0; i < procCount; i++ {
		// Ensuring that each thread only reads the area it is meant to read
		// by making a copy of the area which it gets access to
		diskBytesCopy, err := verify.CopyByteArea(i*(int(bundleSize)), (i+1)*(int(bundleSize)), reader)
		if err != nil {
			return err
		}
		waitGroup.Add(1)
		fmt.Printf("Starting thread %d with blockStart %d and blockEnd %d\n", i, i*(int(bundleSize)), (i+1)*(int(bundleSize)))
		nodeChannels[i] = make(chan verify.Node, int(math.Ceil(bundleSize/2000)))
		go checksumBlock(i*(int(bundleSize)), (i+1)*(int(bundleSize)), int(bundleSize), diskBytesCopy, nodeChannels[i], i, &waitGroup)
	}

	waitGroup.Wait()
	db, err := verify.OpenDB("./fsverify.db", false)
	if err != nil {
		return err
	}

	// All generated nodes are written to the database at once
	// while this is worse for the speed of verifysetup.
	// it ensures that no write conflicts happen
	// which could be caused by multiple threads accessing the same database
	for i := 0; i < procCount; i++ {
		channel := nodeChannels[i]
		err = db.Batch(func(tx *bolt.Tx) error {
			for j := 0; j < int(blockCount); j++ {
				err := core.AddNode(<-channel, tx)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	signature, err := core.SignDatabase("./fsverify.db", minisignDir)
	if err != nil {
		return err
	}
	sig := minisign.Signature{}
	err = sig.UnmarshalText(signature)
	if err != nil {
		return err
	}

	// The untrusted Signature is stored in a special way
	// requiring special decoding of it to represent it as a string
	var UntrustedSignature [2 + 8 + ed25519.SignatureSize]byte
	binary.LittleEndian.PutUint16(UntrustedSignature[:2], sig.Algorithm)
	binary.LittleEndian.PutUint64(UntrustedSignature[2:10], sig.KeyID)
	copy(UntrustedSignature[10:], sig.Signature[:])
	unsignedHash := base64.StdEncoding.EncodeToString(UntrustedSignature[:])
	signedHash := base64.StdEncoding.EncodeToString(sig.CommentSignature[:])

	fsverifydb, err := os.Open("./fsverify.db")
	if err != nil {
		return err
	}
	defer db.Close()
	dbInfo, err := fsverifydb.Stat()
	if err != nil {
		return err
	}
	dbSize := dbInfo.Size()
	verifyPart := make([]byte, 200+dbSize)
	header, err := core.CreateHeader(unsignedHash, signedHash, int(diskSize), int(dbSize))
	database := make([]byte, dbSize)
	_, err = fsverifydb.Read(database)
	if err != nil {
		return err
	}

	copy(verifyPart, header)
	copy(verifyPart[200:], database)

	verifyfs, err := os.Create(args[2])
	if err != nil {
		return err
	}
	defer verifyfs.Close()
	_, err = verifyfs.Write(verifyPart)
	return err
}
