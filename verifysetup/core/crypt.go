package core

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

func CalculateBlockHash(block []byte) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(block)); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:32]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}
