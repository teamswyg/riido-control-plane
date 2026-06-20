package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func loadJSONFile[T any](path string) (T, error) {
	var value T
	file, err := os.Open(path)
	if err != nil {
		return value, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return value, fmt.Errorf("%s must contain one JSON document: %w", path, err)
	}
	return value, nil
}
