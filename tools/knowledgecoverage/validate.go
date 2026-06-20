package main

import (
	"fmt"
	"os"
)

const manifestSchema = "riido-executable-knowledge-coverage.v1"

func validateManifest(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "schema_version must be "+manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, title, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.ScanRoots) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "scan_roots and assertions must not be empty")
	}
	if !completeLoop(m.Loop) {
		problems = append(problems, "loop must define observe/hypothesis/execute/evaluate/retrospective")
	}
	for _, path := range append([]string{m.Workflow}, m.ScanRoots...) {
		if _, err := os.Stat(resolvePath(root, path)); err != nil {
			problems = append(problems, fmt.Sprintf("missing path %q", path))
		}
	}
	return append(problems, validateManualGroups(m)...)
}

func validateManualGroups(m manifest) []string {
	var problems []string
	seen := map[string]bool{}
	for _, group := range m.ManualGroups {
		if group.ID == "" || group.Owner == "" || group.Reason == "" || group.NextArtifact == "" {
			problems = append(problems, "manual group id, owner, reason, and next_artifact are required")
		}
		if seen[group.ID] {
			problems = append(problems, fmt.Sprintf("duplicate manual group %q", group.ID))
		}
		seen[group.ID] = true
		if len(group.Paths) == 0 && len(group.PathPrefixes) == 0 {
			problems = append(problems, fmt.Sprintf("manual group %q has no paths or path_prefixes", group.ID))
		}
	}
	return problems
}
