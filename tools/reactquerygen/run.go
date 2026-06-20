package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func run(openAPIPath, outPath string) error {
	if strings.TrimSpace(openAPIPath) == "" || strings.TrimSpace(outPath) == "" {
		return errors.New("usage: go run ./tools/reactquerygen -openapi <path> -out <path>")
	}
	spec, err := loadOpenAPI(openAPIPath)
	if err != nil {
		return err
	}
	coreBody, err := generate(spec)
	if err != nil {
		return err
	}
	reactBody, err := generateReact(spec)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, coreBody, 0o644); err != nil {
		return err
	}
	return os.WriteFile(reactOutPath(outPath), reactBody, 0o644)
}

func reactOutPath(outPath string) string {
	return strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".react.ts"
}

func loadOpenAPI(path string) (openAPISpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return openAPISpec{}, fmt.Errorf("read %s: %w", path, err)
	}
	var spec openAPISpec
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&spec); err != nil {
		return openAPISpec{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(spec.Paths) == 0 {
		return openAPISpec{}, errors.New("OpenAPI paths are required")
	}
	return spec, nil
}
