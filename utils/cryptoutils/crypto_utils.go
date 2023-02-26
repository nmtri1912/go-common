package cryptoutils

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

func stringHasher(algorithm hash.Hash, text string) string {
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func SHA256(text string) string {
	algorithm := sha256.New()
	return stringHasher(algorithm, text)
}
