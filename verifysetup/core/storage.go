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
	block := make([]byte, end-start)
	_, err := device.Seek(int64(start), 0)
	if err != nil {
		return []byte{}, err
	}
	_, err = device.Read(block)
	TotalReadBlocks = TotalReadBlocks + (end - start)
	return block, err
}

func CreateNode(blockStart int, blockEnd int, block []byte, prevNode *verify.Node) (verify.Node, error) {
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
		prevNodeHash = "Entrypoint"
	}
	node.PrevNodeSum = prevNodeHash
	return node, nil
}

func AddNode(node verify.Node, db *bolt.DB, dbPath string) error {
	var err error
	var deferDB bool
	if db == nil {
		db, err = bolt.Open(dbPath, 0777, nil)
		if err != nil {
			return err
		}
		deferDB = true
	} else if db.IsReadOnly() {
		return fmt.Errorf("Error: database is opened read only, unable to add nodes")
	}
	err = db.Update(func(tx *bolt.Tx) error {
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

	})
	if deferDB {
		defer db.Close()
	}
	return err
}
