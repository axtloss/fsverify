package core

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"
)

type Header struct {
	MagicNumber    int
	Signature      string
	FilesystemSize int
	TableSize      int
}

type Node struct {
	BlockStart  int
	BlockEnd    int
	BlockSum    string
	PrevNodeSum string
}

func ReadHeader(partition string) (Header, error) {
	_, exist := os.Stat(partition)
	if os.IsNotExist(exist) {
		return Header{}, fmt.Errorf("Cannot find partition %s", partition)
	}
	part, err := os.Open(partition)
	if err != nil {
		return Header{}, err
	}
	defer part.Close()

	header := Header{}
	reader := bufio.NewReader(part)
	MagicNumber := make([]byte, 2)
	Signature := make([]byte, 302)
	FileSystemSize := make([]byte, 4)
	TableSize := make([]byte, 4)

	_, err = reader.Read(MagicNumber)
	MagicNum := binary.BigEndian.Uint16(MagicNumber)
	if MagicNum != 0xACAB { // The Silliest of magic numbers
		return Header{}, err
	}
	header.MagicNumber = int(MagicNum)

	_, err = reader.Read(Signature)
	if err != nil {
		return Header{}, err
	}
	_, err = reader.Read(FileSystemSize)
	if err != nil {
		return Header{}, err
	}
	_, err = reader.Read(TableSize)
	if err != nil {
		return Header{}, err
	}

	header.Signature = string(Signature)
	header.FilesystemSize = int(binary.BigEndian.Uint16(FileSystemSize))
	header.TableSize = int(binary.BigEndian.Uint16(TableSize))
	return header, nil
}

func OpenDB() (*bolt.DB, error) {
	_, exist := os.Stat("my.db") // TODO: use configuration file for db path
	if os.IsNotExist(exist) {
		os.Create("my.db")
	}
	db, err := bolt.Open("my.db", 0777, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AddNode(node Node, db *bolt.DB) error {
	var err error
	var deferDB bool
	if db == nil {
		db, err = OpenDB()
		if err != nil {
			return err
		}
		deferDB = true
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

func GetNode(checksum string, db *bolt.DB) (Node, error) {
	var err error
	var deferDB bool
	if db == nil {
		db, err = OpenDB()
		if err != nil {
			return Node{}, err
		}
		deferDB = true
	}
	var node Node
	err = db.View(func(tx *bolt.Tx) error {
		nodes := tx.Bucket([]byte("Nodes"))
		app := nodes.Get([]byte(checksum))
		err := json.Unmarshal(app, &node)
		return err
	})
	if deferDB {
		defer db.Close()
	}
	return node, err
}
