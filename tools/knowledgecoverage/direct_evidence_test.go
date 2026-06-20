package main

import (
	"strings"
	"testing"
)

func TestScanRejectsDirectSSOTWithoutEvidenceTool(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/direct.md", "# Direct\n")
	mustWrite(t, root, "docs/direct.riido.json", `{"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`)
	m := manifest{ScanRoots: []string{"docs"}}
	_, problems := scanDocs(root, m)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, "\n"), "must declare evidence_tool") {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestScanRejectsDirectSSOTWithoutEvidenceWorkflow(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/direct.md", "# Direct\n")
	mustWrite(t, root, "docs/direct.riido.json", `{"evidence_tool":"tools/directcheck","loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`)
	mustWrite(t, root, "tools/directcheck/main.go", "package main\n")
	mustWrite(t, root, ".github/workflows/direct.yml", "steps:\n  - run: go test ./tools/directcheck\n")
	m := manifest{ScanRoots: []string{"docs"}}
	_, problems := scanDocs(root, m)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, "\n"), "must run check-doc with evidence-out") {
		t.Fatalf("problems = %#v", problems)
	}
}
