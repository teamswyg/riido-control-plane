package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func readStrictJSON(root, path string, value any) error {
	raw, err := readRootFile(root, path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("JSON must contain exactly one value")
	}
	return nil
}

func readJSON(root, path string, value any) error {
	raw, err := readRootFile(root, path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, value)
}

func readRootFile(rootPath, path string) (raw []byte, returnErr error) {
	if filepath.IsAbs(path) || path == "" || filepath.Clean(path) == "." {
		return nil, errors.New("repository path must be a non-empty relative file")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	defer func() { returnErr = errors.Join(returnErr, root.Close()) }()
	return root.ReadFile(filepath.FromSlash(path))
}

func containsAll(value string, items []string) bool {
	for _, item := range items {
		if !strings.Contains(value, item) {
			return false
		}
	}
	return true
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func status(passed bool) string {
	if passed {
		return "passed"
	}
	return "failed"
}
