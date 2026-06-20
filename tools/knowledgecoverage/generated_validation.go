package main

import (
	"fmt"
	"os"
)

func validateGeneratedDocs(root string, docs []docClass) []string {
	workflows := loadWorkflows(root)
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
		if !workflowMentionsTool(workflows, doc.GeneratorTool) {
			problems = append(problems, fmt.Sprintf("generated doc %q generator %q is not referenced by CI workflow", doc.Path, doc.GeneratorTool))
		}
		if !workflowHasGeneratorEvidence(workflows, doc.GeneratorTool) {
			problems = append(problems, fmt.Sprintf("generated doc %q generator %q must run check-doc with evidence-out in CI", doc.Path, doc.GeneratorTool))
		}
	}
	return problems
}
