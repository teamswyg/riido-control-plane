package main

import (
	"fmt"
	"os"
)

func workflowText(root, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("source workflow path is required")
	}
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
