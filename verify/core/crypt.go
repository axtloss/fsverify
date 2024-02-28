package core

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
)

// calculateStringHash calculates the sha1 checksum of a given string a.
func calculateStringHash(a string) (string, error) {
	hash := sha1.New()
	hash.Write([]byte(a))
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}

// CalculateBlockHash calculates the sha1 checksum of a given byte slice b.
func CalculateBlockHash(b []byte) (string, error) {
	hash := sha1.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}
