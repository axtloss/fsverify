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

// Header contains all information stored in the header of a fsverify partition.
type Header struct {
	MagicNumber    int
	Signature      string
	FilesystemSize int
	FilesystemUnit int
	TableSize      int
	TableUnit      int
}

// Node contains all information stored in a database node.
// If the Node is the first node in the database, PrevNodeSum should be set to Entrypoint.
type Node struct {
	BlockStart  int
	BlockEnd    int
	BlockSum    string
	PrevNodeSum string
}

// GetHash returns the hash of all fields of a Node combined.
// The Node fields are combined in the order BlockStart, BlockEnd, BlockSum and PrevNodeSum
func (n *Node) GetHash() (string, error) {
	return calculateStringHash(fmt.Sprintf("%d%d%s%s", n.BlockStart, n.BlockEnd, n.BlockSum, n.PrevNodeSum))
}

// parseUnitSpec parses the file size unit specified in the header and returns it as an according multiplier.
// In the case of an invalid Unit byte the function returns -1.
func parseUnitSpec(size []byte) int {
	switch size[0] {
	case 0:
		return 1
	case 1:
		return 1000
	case 2:
		return 1000 * 1000
	case 3:
		return 1000 * 1000 * 10000
	case 4:
		return 100000000000000
	case 5:
		return 1000000000000000
	default:
		return -1
	}
}

// ReadHeader reads the partition header and puts it in a variable of type Header.
// If any field fails to be read, the function returns an empty Header struct and the error.
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
	// Since the size of each field is already known
	// it is best to hard code them, in the case
	// that a field goes over its allocated size
	// fsverify should (and will) fail
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

// ReadDB reads the database from a fsverify partition.
// It verifies the the size of the database with the size specified in the partition header and returns an error if the sizes do not match.
// Due to limitations with bbolt the database gets written to a temporary path and the function returns the path to the database.
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

	// The area taken up by the header
	// it is useless for this reader instance
	// and will be skipped completely
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

	// Reading the specified table size allows for tamper protection
	// in the case that the partition was tampered with "lazily"
	// meaning that only the database was modified, and not the header
	// if that is the case, the database would be lacking data, making it unusable
	db := make([]byte, header.TableSize*header.TableUnit)
	n, err := io.ReadFull(reader, db)
	if err != nil {
		return "", err
	}
	if n != header.TableSize*header.TableUnit {
		return "", fmt.Errorf("Database is not expected size. Expected %d, got %d", header.TableSize*header.TableUnit, n)
	}
	fmt.Printf("db: %d\n", n)

	// Write the database to a temporary directory
	// to ensure that it disappears after the next reboot
	temp, err := os.MkdirTemp("", "*-fsverify")
	if err != nil {
		return "", err
	}

	// The file permission is immediately set to 0700
	// this ensures that the database is not modified
	// after it has been written
	err = os.WriteFile(temp+"/verify.db", db, 0700)
	if err != nil {
		return "", err
	}

	return temp + "/verify.db", nil
}

// OpenDB opens a bbolt database and returns a bbolt instance.
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

// GetNode retrieves a Node from the database based on the hash identifier.
// If db is set to nil, the function will open the database in read-only mode itself.
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

// CopyByteArea copies an area of bytes from a reader.
// It verifies that the reader reads the wanted amount of bytes, and returns an error if this is not the case.
func CopyByteArea(start int, end int, reader *bytes.Reader) ([]byte, error) {
	if end-start < 0 {
		return []byte{}, fmt.Errorf("tried creating byte slice with negative length. %d to %d total %d\n", start, end, end-start)
	} else if end-start > 2000 {
		return []byte{}, fmt.Errorf("tried creating byte slice with length over 2000. %d to %d total %d\n", start, end, end-start)
	}
	bytes := make([]byte, end-start)
	n, err := reader.ReadAt(bytes, int64(start))
	if err != nil {
		return nil, err
	} else if n != end-start {
		return nil, fmt.Errorf("Unable to read requested size. Expected %d, got %d", end-start, n)
	}
	return bytes, nil
}
