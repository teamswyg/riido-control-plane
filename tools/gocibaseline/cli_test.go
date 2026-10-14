package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoCIBaselineMainRunWritesDocAndEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n")
	writeFile(t, filepath.Join(repo, ".github/workflows/ci.yml"), "go test ./...\n")
	manifestPath := filepath.Join(repo, "baseline.json")
	manifestBody, err := json.Marshal(manifest{
		SchemaVersion: manifestSchema,
		ID:            "go-ci",
		Title:         "Go CI",
		GeneratedDoc:  "docs/go-ci.md",
		Workflow:      ".github/workflows/ci.yml",
		Evidence:      "go-ci-evidence",
		Gates: []gate{{
			ID: "go-test", Summary: "Go tests", Contains: []string{"go test ./..."},
		}},
		Loop: completeLoop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, manifestPath, string(manifestBody))
	evidenceOut := filepath.Join(repo, "out/evidence.json")
	err = mainRun([]string{
		"-repo", repo, "-manifest", "baseline.json",
		"-write-doc", "-evidence-out", evidenceOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(repo, "docs/go-ci.md"))
	assertExists(t, evidenceOut)
}

func TestGoCIBaselineMainRunRejectsBadFlag(t *testing.T) {
	if err := mainRun([]string{"-unknown"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func completeLoop() loopRecord {
	return loopRecord{
		Observation:   "observe",
		Hypothesis:    "hypothesis",
		Execute:       "execute",
		Evaluate:      "evaluate",
		Retrospective: "retrospective",
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
