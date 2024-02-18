package core

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"aead.dev/minisign"
	"github.com/axtloss/fsverify/config"
	"github.com/tarm/serial"
)

// fileReadKey reads the public minisign key from a file specified in config.KeyLocation.
func fileReadKey() (string, error) {
	if _, err := os.Stat(config.KeyLocation); os.IsNotExist(err) {
		return "", fmt.Errorf("Key location %s does not exist", config.KeyLocation)
	}
	file, err := os.Open(config.KeyLocation)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// A public key is never longer than 56 bytes
	key := make([]byte, 56)
	reader := bufio.NewReader(file)
	n, err := reader.Read(key)
	if n != 56 {
		return "", fmt.Errorf("Key does not match expected key size. Expected 56, got %d", n)
	}
	if err != nil {
		return "", err
	}
	return string(key), nil
}

// serialReadKey reads the public minisign key from a usb tty specified in config.KeyLocation.
func serialReadKey() (string, error) {
	// Since the usb serial is tested with an arduino
	// it is assumed that the tty device does not always exist
	// and can be manually plugged in by the user
	if _, err := os.Stat(config.KeyLocation); !os.IsNotExist(err) {
		fmt.Println("Reconnect arduino now")
		for true {
			if _, err := os.Stat(config.KeyLocation); os.IsNotExist(err) {
				break
			}
		}
	} else {
		fmt.Println("Connect arduino now")
	}
	for true {
		if _, err := os.Stat(config.KeyLocation); !os.IsNotExist(err) {
			break
		}
	}
	fmt.Println("Arduino connected")
	c := &serial.Config{Name: config.KeyLocation, Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		return "", err
	}

	key := ""
	for true {
		buf := make([]byte, 128)
		n, err := s.Read(buf)
		if err != nil {
			return "", err
		}
		defer s.Close()
		key = key + fmt.Sprintf("%q", buf[:n])
		// ensure that two tab sequences are read
		// meaning that the entire key has been captured
		// since the key is surrounded by a tab sequence
		if strings.Count(key, "\\t") == 2 {
			break
		}
	}
	key = strings.ReplaceAll(key, "\\t", "")
	key = strings.ReplaceAll(key, "\"", "")
	if len(key) != 56 {
		return "", fmt.Errorf("Key does not match expected key size. Expected 56, got %d", len(key))
	}
	return key, nil
}

// ReadKey is a wrapper function to call the proper readKey function according to config.KeyStore.
func ReadKey() (string, error) {
	switch config.KeyStore {
	case 0:
		return fileReadKey()
	case 1:
		return fileReadKey()
	case 2:
		return "", nil // TPM
	case 3:
		return serialReadKey()
	}
	return "", nil
}

// ReadBlock reads a data area of a bytes.Reader specified in the given node.
// It additionally verifies that the amount of bytes read equal the wanted amount and returns an error if this is not the case.
func ReadBlock(node Node, part *bytes.Reader, totalReadBlocks int) ([]byte, int, error) {
	if node.BlockEnd-node.BlockStart < 0 {
		return []byte{}, -1, fmt.Errorf("tried creating byte slice with negative length. %d to %d total %d\n", node.BlockStart, node.BlockEnd, node.BlockEnd-node.BlockStart)
	} else if node.BlockEnd-node.BlockStart > 2000 {
		return []byte{}, -1, fmt.Errorf("tried creating byte slice with length over 2000. %d to %d total %d\n", node.BlockStart, node.BlockEnd, node.BlockEnd-node.BlockStart)
	}
	block := make([]byte, node.BlockEnd-node.BlockStart)
	blockSize := node.BlockEnd - node.BlockStart
	_, err := part.Seek(int64(node.BlockStart), 0)
	if err != nil {
		return []byte{}, -1, err
	}
	n, err := part.Read(block)
	if err != nil {
		return block, -1, err
	} else if n != blockSize {
		return block, -1, fmt.Errorf("Did not read correct amount of bytes. Expected: %d, Got: %d", blockSize, n)
	}
	return block, totalReadBlocks + 1, err
}

// VerifySignature verifies the database using a given signature and public key.
func VerifySignature(key string, signature string, database string) (bool, error) {
	var pk minisign.PublicKey
	if err := pk.UnmarshalText([]byte(key)); err != nil {
		return false, err
	}

	data, err := os.ReadFile(database)
	if err != nil {
		return false, err
	}

	return minisign.Verify(pk, data, []byte(signature)), nil
}

// VerifyBlock verifies a byte slice with the hash in a given Node.
func VerifyBlock(block []byte, node Node) error {
	calculatedBlockHash, err := CalculateBlockHash(block)
	if err != nil {
		return err
	}
	wantedBlockHash := node.BlockSum
	if strings.Compare(calculatedBlockHash, strings.TrimSpace(wantedBlockHash)) == 0 {
		return nil
	}
	return fmt.Errorf("Node %s ranging from %d to %d does not match block. Expected %s, got %s.", node.PrevNodeSum, node.BlockStart, node.BlockEnd, wantedBlockHash, calculatedBlockHash)
}

// VerifyNode verifies that the current Node is valid by matching the checksum of it with the PrevNodeSum field of the next node.
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
