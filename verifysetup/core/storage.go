package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	verify "github.com/axtloss/fsverify/verify/core"
	bolt "go.etcd.io/bbolt"
)

// ReadBlock reads the bytes in a specified ranges from a bytes.Reader.
// It additionally verifies that the amount of bytes read match with the size of the area and fails if the they do not match.
func ReadBlock(start int, end int, device *bytes.Reader) ([]byte, error) {
	if end-start < 0 {
		return []byte{}, fmt.Errorf("tried creating byte slice with negative length. %d to %d total %d\n", start, end, end-start)
	}
	block := make([]byte, end-start)
	_, err := device.Seek(int64(start), 0)
	if err != nil {
		return []byte{}, err
	}
	_, err = device.Read(block)
	return block, err
}

// CreateNode creates a Node based on given parameters.
// If prevNode is set to nil, meaning this node is the first node in a verification chain, prevNodeHash is set to "EntrypointN" with N being the number of entrypoint.
func CreateNode(blockStart int, blockEnd int, block []byte, prevNode *verify.Node, n string) (verify.Node, error) {
	node := verify.Node{}
	node.BlockStart = blockStart
	node.BlockEnd = blockEnd
	blockHash, err := CalculateBlockHash(block)
	if err != nil {
		return verify.Node{}, err
	}
	node.BlockSum = blockHash
	var prevNodeHash string
	if prevNode.PrevNodeSum != "" {
		prevNodeHash, err = prevNode.GetHash()
		if err != nil {
			return verify.Node{}, err
		}
	} else {
		prevNodeHash = "Entrypoint" + n
	}
	node.PrevNodeSum = prevNodeHash
	return node, nil
}

// AddNode adds a node to the bucket "Nodes" in the database.
// It assumes that a database transaction has already been started and takes bolt.Tx as an argument.
func AddNode(node verify.Node, tx *bolt.Tx) error {
	if node.BlockStart == node.BlockEnd {
		return nil
	}
	nodes, err := tx.CreateBucketIfNotExists([]byte("Nodes"))
	if err != nil {
		return err
	}
	if buf, err := json.Marshal(node); err != nil {
		return err
	} else if err := nodes.Put([]byte(node.PrevNodeSum), buf); err != nil {
		return err
	}
	return nil
}

// CreateHeader creates a header to be used in an fsverify partition containing all necessary information.
func CreateHeader(unsignedHash string, signedHash string, diskSize int, tableSize int) ([]byte, error) {
	header := make([]byte, 200)
	header[0] = 0xAC
	header[1] = 0xAB
	copy(header[2:], []byte(unsignedHash))
	copy(header[102:], []byte(signedHash))

	disk := make([]byte, 4)
	binary.BigEndian.PutUint32(disk, uint32(diskSize))
	copy(header[190:], disk)

	db := make([]byte, 4)
	binary.BigEndian.PutUint32(db, uint32(tableSize))
	copy(header[195:], db)

	return header, nil
}
