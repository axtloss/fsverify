package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	verify "github.com/axtloss/fsverify/core"
	bolt "go.etcd.io/bbolt"
)

var TotalReadBlocks = 0

func ReadBlock(start int, end int, device *bytes.Reader) ([]byte, error) {
	if end-start < 0 {
		return []byte{}, fmt.Errorf("ERROR: tried creating byte slice with negative length. %d to %d total %d\n", start, end, end-start)
	} else if end-start > 2000 {
		return []byte{}, fmt.Errorf("ERROR: tried creating byte slice with length over 2000. %d to %d total %d\n", start, end, end-start)
	}
	block := make([]byte, end-start)
	_, err := device.Seek(int64(start), 0)
	if err != nil {
		return []byte{}, err
	}
	_, err = device.Read(block)
	TotalReadBlocks = TotalReadBlocks + (end - start)
	return block, err
}

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
