package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

func calculateStringHash(a string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(a))
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}

func calculateFileHash(file *os.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:20]
	return strings.TrimSpace(fmt.Sprintf("%x", hashInBytes)), nil
}
