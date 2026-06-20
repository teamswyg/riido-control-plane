package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateGeneratedDocs(root string, docs []docClass) []string {
	workflowText := allWorkflowText(root)
	var problems []string
	for _, doc := range docs {
		if doc.Kind != "generated" {
			continue
		}
		if doc.GeneratorTool == "" {
			problems = append(problems, fmt.Sprintf("generated doc %q must name a tools/<name> generator", doc.Path))
			continue
		}
		if _, err := os.Stat(resolvePath(root, doc.GeneratorTool)); err != nil {
			problems = append(problems, fmt.Sprintf("generated doc %q references missing generator %q", doc.Path, doc.GeneratorTool))
		}
		if !strings.Contains(workflowText, doc.GeneratorTool) {
			problems = append(problems, fmt.Sprintf("generated doc %q generator %q is not referenced by CI workflow", doc.Path, doc.GeneratorTool))
		}
	}
	return problems
}

func allWorkflowText(root string) string {
	var b strings.Builder
	_ = filepath.WalkDir(resolvePath(root, ".github/workflows"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return err
		}
		data, err := os.ReadFile(path)
		if err == nil {
			b.Write(data)
			b.WriteByte('\n')
		}
		return nil
	})
	return b.String()
}
