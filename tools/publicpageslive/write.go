package main

import (
	"encoding/json"
	"io"
	"os"
)

func write(path string, stdout io.Writer, rec record) error {
	body, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if path == "" || path == "-" {
		_, err = stdout.Write(body)
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
