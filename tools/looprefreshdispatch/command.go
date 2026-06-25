package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseRefreshWorkflowCommand(root, command string) (string, error) {
	fields := strings.Fields(command)
	if len(fields) != 6 {
		return "", fmt.Errorf("unsupported refresh command shape")
	}
	if fields[0] != "gh" || fields[1] != "workflow" || fields[2] != "run" ||
		fields[4] != "--ref" || fields[5] != "main" {
		return "", fmt.Errorf("unsupported refresh workflow command")
	}
	workflow := fields[3]
	if !safeWorkflowFile(workflow) {
		return "", fmt.Errorf("unsafe workflow file %q", workflow)
	}
	if err := workflowExists(root, workflow); err != nil {
		return "", err
	}
	return workflow, nil
}

func safeWorkflowFile(value string) bool {
	if strings.ContainsAny(value, `/\`) {
		return false
	}
	if !strings.HasSuffix(value, ".yml") && !strings.HasSuffix(value, ".yaml") {
		return false
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		if r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return value != ".yml" && value != ".yaml"
}

func workflowExists(root, workflow string) error {
	path := filepath.Join(root, ".github", "workflows", workflow)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("workflow %q is not present: %w", workflow, err)
	}
	if info.IsDir() {
		return fmt.Errorf("workflow %q is a directory", workflow)
	}
	return nil
}
