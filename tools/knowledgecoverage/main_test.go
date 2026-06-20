package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := run("../..", "docs/executable-knowledge.riido.json", out, false, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "verified" || got.ScannedCount == 0 || got.ManualCount != 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedToolCount || len(got.GeneratedMissingWorkflow) != 0 {
		t.Fatalf("generated tool coverage drifted: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedEvidenceWorkflowCount || len(got.GeneratedMissingEvidenceWorkflow) != 0 {
		t.Fatalf("generated evidence workflow coverage drifted: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedArtifactBindingCount || len(got.GeneratedMissingArtifactBinding) != 0 {
		t.Fatalf("generated artifact binding coverage drifted: %+v", got)
	}
	if got.DirectSSOTCount != got.DirectLoopCount || len(got.DirectMissingLoop) != 0 {
		t.Fatalf("direct SSOT loop coverage drifted: %+v", got)
	}
}

func TestScanRejectsUnregisteredManualDoc(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/known.md", "# Known\n")
	m := manifest{ScanRoots: []string{"docs"}}
	_, problems := scanDocs(root, m)
	if len(problems) == 0 || !strings.Contains(problems[0], "unregistered manual") {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestScanFilesIncludeRootMarkdown(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "README.md", "# Manual\n")
	m := manifest{ScanFiles: []string{"README.md"}}
	docs, problems := scanDocs(root, m)
	if len(docs) != 1 || docs[0].Path != "README.md" {
		t.Fatalf("docs = %#v", docs)
	}
	if len(problems) == 0 || !strings.Contains(problems[0], "unregistered manual") {
		t.Fatalf("problems = %#v", problems)
	}
}

func mustWrite(t *testing.T, root, path, text string) {
	t.Helper()
	if err := writeText(filepath.Join(root, filepath.FromSlash(path)), text); err != nil {
		t.Fatal(err)
	}
}
