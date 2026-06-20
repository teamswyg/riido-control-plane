package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func fileHashes(paths map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for name, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		out[name] = hex.EncodeToString(sum[:])
	}
	return out, nil
}
