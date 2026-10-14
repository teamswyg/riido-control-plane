package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func writeEvidence(path string, result evidence) (returnErr error) {
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, root.Close()) }()
	file, err := root.OpenFile(filepath.Base(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, file.Close()) }()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	_, err = file.Write(append(raw, '\n'))
	return err
}
