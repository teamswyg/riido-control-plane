package main

import (
	"os"
	"path/filepath"
	"strings"
)

type workflowDoc struct {
	Path string
	Text string
}

func loadWorkflows(root string) []workflowDoc {
	var workflows []workflowDoc
	_ = filepath.WalkDir(resolvePath(root, ".github/workflows"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isWorkflowPath(path) {
			return err
		}
		data, err := os.ReadFile(path)
		if err == nil {
			workflows = append(workflows, workflowDoc{Path: slashPath(root, path), Text: string(data)})
		}
		return nil
	})
	return workflows
}

func isWorkflowPath(path string) bool {
	return strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")
}
