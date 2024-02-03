package core

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	bolt "go.etcd.io/bbolt"
)

type Header struct {
	MagicNumber    int
	Signature      string
	FilesystemSize int
	FilesystemUnit int
	TableSize      int
	TableUnit      int
}

type Node struct {
	BlockStart  int
	BlockEnd    int
	BlockSum    string
	PrevNodeSum string
}

func parseUnitSpec(size []byte) int {
	switch size[0] {
	case 0:
		return 1
	case 1:
		return 1000
	case 2:
		return 1000000
	case 3:
		return 1000000000
	case 4:
		return 1000000000000
	case 5:
		return 1000000000000000
	default:
		return -1
	}
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
	UntrustedHash := make([]byte, 100)
	TrustedHash := make([]byte, 88)
	FilesystemSize := make([]byte, 4)
	FilesystemUnit := make([]byte, 1)
	TableSize := make([]byte, 4)
	TableUnit := make([]byte, 1)

	_, err = reader.Read(MagicNumber)
	MagicNum := binary.BigEndian.Uint16(MagicNumber)
	if MagicNum != 0xACAB { // The Silliest of magic numbers
		return Header{}, err
	}
	header.MagicNumber = int(MagicNum)

	_, err = reader.Read(UntrustedHash)
	if err != nil {
		return Header{}, err
	}
	_, err = reader.Read(TrustedHash)
	if err != nil {
		return Header{}, err
	}
	_, err = reader.Read(FilesystemSize)
	if err != nil {
		return Header{}, err
	}
	_, err = reader.Read(FilesystemUnit)
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

	header.Signature = fmt.Sprintf("untrusted comment: signature from minisign secret key\r\n%s\r\ntrusted comment: timestamp:0\tfile:fsverify\thashed\r\n%s\r\n", UntrustedHash, TrustedHash)
	header.FilesystemSize = int(binary.BigEndian.Uint16(FilesystemSize))
	header.TableSize = int(binary.BigEndian.Uint32(TableSize))
	header.FilesystemUnit = parseUnitSpec(FilesystemUnit)
	header.TableUnit = parseUnitSpec(TableUnit)
	if header.FilesystemUnit == -1 || header.TableUnit == -1 {
		return Header{}, fmt.Errorf("Error: unit size for Filesystem or Table invalid: fs: %x, table: %x", FilesystemUnit, TableUnit)
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

	_, err = reader.Read(make([]byte, 200))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	header, err := ReadHeader(partition)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	db := make([]byte, header.TableSize*header.TableUnit)
	_, err = io.ReadFull(reader, db)
	if err != nil {
		return "", err
	}

	temp, err := os.MkdirTemp("", "*-fsverify")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(temp+"/verify.db", db, 0700)
	if err != nil {
		return "", err
	}

	return temp + "/verify.db", nil
}

func OpenDB(dbpath string, readonly bool) (*bolt.DB, error) {
	_, exist := os.Stat(dbpath)
	if os.IsNotExist(exist) {
		os.Create(dbpath)
	}
	db, err := bolt.Open(dbpath, 0777, &bolt.Options{ReadOnly: readonly})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AddNode(node Node, db *bolt.DB) error {
	var err error
	var deferDB bool
	if db == nil {
		db, err = OpenDB("my.db", false)
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

func GetNode(checksum string, db *bolt.DB) (Node, error) {
	var err error
	var deferDB bool
	if db == nil {
		db, err = OpenDB("my.db", true)
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
