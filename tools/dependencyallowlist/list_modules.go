package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func listModules(dir string) ([]goModule, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("go list -m -json all: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("go list -m -json all: %w", err)
	}
	return decodeModules(output)
}

func decodeModules(output []byte) ([]goModule, error) {
	decoder := json.NewDecoder(bytes.NewReader(output))
	var modules []goModule
	for {
		var module goModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode go list module: %w", err)
		}
		modules = append(modules, module)
	}
	return modules, nil
}
