package core

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
)

func calculateStringHash(a string) (string, error) {
	hash := sha1.New()
	hash.Write([]byte(a))
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}

func CalculateBlockHash(block []byte) (string, error) {
	hash := sha1.New()
	if _, err := io.Copy(hash, bytes.NewReader(block)); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}
