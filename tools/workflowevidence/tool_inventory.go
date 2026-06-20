package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func evidenceToolDirs(root string) []string {
	entries, err := os.ReadDir(repoPath(root, "tools"))
	if err != nil {
		return nil
	}
	var tools []string
	for _, entry := range entries {
		if entry.IsDir() && toolSupportsEvidenceOut(root, entry.Name()) {
			tools = append(tools, entry.Name())
		}
	}
	sort.Strings(tools)
	return tools
}

func toolSupportsEvidenceOut(root, tool string) bool {
	files, err := filepath.Glob(repoPath(root, filepath.Join("tools", tool, "*.go")))
	if err != nil {
		return false
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err == nil && strings.Contains(string(data), "evidence-out") {
			return true
		}
	}
	return false
}

func workflowCallsEvidenceTool(workflowTexts []string, tool string) bool {
	needle := "./tools/" + tool
	for _, text := range workflowTexts {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
