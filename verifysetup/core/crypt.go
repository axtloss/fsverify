package core

import (
	"aead.dev/minisign"
	"bytes"
	"crypto/sha1"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
)

// CalculateBlockHash calculates the sha1 checksum of a byte slice.
func CalculateBlockHash(block []byte) (string, error) {
	hash := sha1.New()
	if _, err := io.Copy(hash, bytes.NewReader(block)); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}

// SignDatabase generates a minisign signature of the database using given keys.
// The minisign signature uses "fsverify" as the comments to ensure predictability when fsverify verifies the signature.
func SignDatabase(database string, minisignKeys string) ([]byte, error) {
	fmt.Print("Enter your password (will not echo): ")
	p, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Println("\nSigning database")
	privateKey, err := minisign.PrivateKeyFromFile(string(p), minisignKeys+"/minisign.key")
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(database)
	if err != nil {
		return nil, err
	}
	signature := minisign.SignWithComments(privateKey, data, "fsverify", "fsverify")
	return signature, err
}
