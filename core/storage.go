package core

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	bolt "go.etcd.io/bbolt"
)

type Header struct {
	MagicNumber    int
	Signature      string
	FilesystemSize int
	TableSize      int
	TableUnit      int
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
	TableUnit := make([]byte, 1)

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
	_, err = reader.Read(TableUnit)
	if err != nil {
		return Header{}, err
	}

	header.Signature = string(Signature)
	header.FilesystemSize = int(binary.BigEndian.Uint16(FileSystemSize))
	header.TableSize = int(binary.BigEndian.Uint32(TableSize))
	switch TableUnit[0] {
	case 0:
		header.TableUnit = 1
	case 1:
		header.TableUnit = 1000
	case 2:
		header.TableUnit = 1000000
	case 3:
		header.TableUnit = 1000000000
	case 4:
		header.TableUnit = 1000000000000
	case 5:
		header.TableUnit = 1000000000000000
	default:
		return Header{}, fmt.Errorf("Unknown TableUnit %d", TableUnit)
	}
	return header, nil
}

func ReadDB(partition string) (string, error) {
	_, exist := os.Stat(partition)
	if os.IsNotExist(exist) {
		return "", fmt.Errorf("Cannot find partition %s", partition)
	}
	part, err := os.Open(partition)
	if err != nil {
		return "", err
	}
	defer part.Close()
	reader := bufio.NewReader(part)

	_, err = reader.Read(make([]byte, 313))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	header, err := ReadHeader(partition)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Header:")
	fmt.Println(header.TableSize)
	fmt.Println(header.TableUnit)
	db := make([]byte, header.TableSize*header.TableUnit)
	_, err = reader.Read(db)
	if err != nil {
		return "", err
	}

	temp, err := os.MkdirTemp("", "*-fsverify")
	if err != nil {
		return "", err
	}

	fmt.Println("DB Path:")
	fmt.Println(temp)
	fmt.Println()
	err = os.WriteFile(temp+"/verify.db", db, 0777)
	if err != nil {
		return "", err
	}

	//defer os.RemoveAll(temp)
	return temp + "/verify.db", err
}

func OpenDB(dbpath string) (*bolt.DB, error) {
	_, exist := os.Stat(dbpath)
	if os.IsNotExist(exist) {
		os.Create(dbpath)
	}
	db, err := bolt.Open(dbpath, 0777, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AddNode(node Node, db *bolt.DB) error {
	var err error
	var deferDB bool
	if db == nil {
		db, err = OpenDB("my.db")
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
		db, err = OpenDB("my.db")
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

func VerifyNode(node Node, nextNode Node) error {
	nodeHash, err := calculateStringHash(fmt.Sprintf("%d%d%s%s", node.BlockStart, node.BlockEnd, node.BlockSum, node.PrevNodeSum))
	if err != nil {
		return err
	}
	if strings.Compare(nodeHash, nextNode.PrevNodeSum) != 0 {
		return fmt.Errorf("Node %s is not valid!", node.PrevNodeSum)
	}
	return nil
}
