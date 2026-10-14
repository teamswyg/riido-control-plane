package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func testTokenSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
