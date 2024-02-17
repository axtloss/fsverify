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

//var TotalReadBlocks int = 0

func fileReadKey() (string, error) {
	if _, err := os.Stat(config.KeyLocation); os.IsNotExist(err) {
		return "", fmt.Errorf("Key location %s does not exist", config.KeyLocation)
	}
	file, err := os.Open(config.KeyLocation)
	if err != nil {
		return "", err
	}
	defer file.Close()
	key := make([]byte, 56)
	reader := bufio.NewReader(file)
	n, err := reader.Read(key)
	if n != 56 {
		return "", fmt.Errorf("Error: Key does not match expected key size. expected 56, got %d", n)
	}
	if err != nil {
		return "", err
	}
	return string(key), nil
}

func serialReadKey() (string, error) {
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
		if strings.Count(key, "\\t") == 2 {
			break
		}
	}
	key = strings.ReplaceAll(key, "\\t", "")
	key = strings.ReplaceAll(key, "\"", "")
	if len(key) != 56 {
		return "", fmt.Errorf("Error: Key does not match expected key size. expected 56, got %d", len(key))
	}
	return key, nil
}

func ReadKey() (string, error) {
	switch config.KeyStore {
	case 0:
		return fileReadKey()
	case 1:
		return fileReadKey()
	case 2:
		return "", nil
	case 3:
		return serialReadKey()
	}
	return "", nil
}

func ReadBlock(node Node, part *bytes.Reader, totalReadBlocks int) ([]byte, int, error) {
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

func VerifyBlock(block []byte, node Node) error {
	calculatedBlockHash, err := CalculateBlockHash(block)
	if err != nil {
		return err
	}
	wantedBlockHash := node.BlockSum
	if strings.Compare(calculatedBlockHash, strings.TrimSpace(wantedBlockHash)) == 0 {
		return nil
	}
	return fmt.Errorf("Error: Node %s ranging from %d to %d does not match block. Expected %s, got %s.", node.PrevNodeSum, node.BlockStart, node.BlockEnd, wantedBlockHash, calculatedBlockHash)
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
