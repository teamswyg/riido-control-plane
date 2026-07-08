package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testLoopEvidence() loopEvidence {
	return loopEvidence{
		Observation:   "observe",
		Hypothesis:    "hypothesis",
		Execute:       "execute",
		Evaluate:      "evaluate",
		Retrospective: "retrospective",
	}
}

func testEntry(name string) entry {
	return entry{
		Name:        name,
		Default:     "unset",
		Owner:       "owner",
		Sensitivity: "public",
		Meaning:     "meaning",
	}
}

func testManifest(sourceDir string, entries ...entry) manifest {
	return manifest{
		SchemaVersion: "riido-config-reference.v1",
		ID:            "test-config",
		Title:         "Test Config",
		RiidoTask:     "RIID-TEST",
		GeneratedDoc:  "docs/generated.md",
		Workflow:      "workflow",
		Evidence:      "evidence",
		SourceDir:     sourceDir,
		Entries:       entries,
		Loop:          testLoopEvidence(),
	}
}

func writeMiniRepo(t *testing.T, body string) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	sourceDir := filepath.Join(repo, "cmd", "app")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte(body), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return repo
}

func writeManifestFile(t *testing.T, path string, m manifest) {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}
