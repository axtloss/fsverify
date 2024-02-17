package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"io"
	"os"
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

func (n *Node) GetHash() (string, error) {
	return calculateStringHash(fmt.Sprintf("%d%d%s%s", n.BlockStart, n.BlockEnd, n.BlockSum, n.PrevNodeSum))
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

	header.Signature = fmt.Sprintf("untrusted comment: fsverify\n%s\ntrusted comment: fsverify\n%s\n", string(UntrustedHash), string(TrustedHash))
	header.FilesystemSize = int(binary.BigEndian.Uint32(FilesystemSize))
	header.TableSize = int(binary.BigEndian.Uint32(TableSize))
	header.FilesystemUnit = parseUnitSpec(FilesystemUnit)
	header.TableUnit = parseUnitSpec(TableUnit)
	if header.FilesystemUnit == -1 || header.TableUnit == -1 {
		return Header{}, fmt.Errorf("unit size for Filesystem or Table invalid: fs: %x, table: %x", FilesystemUnit, TableUnit)
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
	n, err := io.ReadFull(reader, db)
	if err != nil {
		fmt.Println("failed reading db")
		fmt.Println(header.TableSize * header.TableUnit)
		return "", err
	}
	if n != header.TableSize*header.TableUnit {
		return "", fmt.Errorf("Database is not expected size. Expected %d, got %d", header.TableSize*header.TableUnit, n)
	}
	fmt.Printf("db: %d\n", n)

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

func CopyByteArea(start int, end int, reader *bytes.Reader) ([]byte, error) {
	bytes := make([]byte, end-start)
	n, err := reader.ReadAt(bytes, int64(start))
	if err != nil {
		return nil, err
	} else if n != end-start {
		return nil, fmt.Errorf("Unable to read requested size. Expected %d, got %d", end-start, n)
	}
	return bytes, nil
}
