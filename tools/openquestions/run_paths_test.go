package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesDocAndEvidenceInTempRepo(t *testing.T) {
	repo := openQuestionsTempRepo(t, validOpenQuestionsManifest())
	evidenceOut := filepath.Join(repo, "out", "evidence.json")

	err := run(options{
		Repo:        repo,
		Manifest:    "open.riido.json",
		WriteDoc:    true,
		CheckDoc:    true,
		EvidenceOut: evidenceOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(repo, "docs", "open.md"), "Open question registry")
	assertFileContains(t, evidenceOut, "control-plane-open-questions")
}

func TestRunRejectsStaleGeneratedDoc(t *testing.T) {
	repo := openQuestionsTempRepo(t, validOpenQuestionsManifest())
	docPath := filepath.Join(repo, "docs", "open.md")
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docPath, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(options{Repo: repo, Manifest: "open.riido.json", CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "is stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}

func TestRunRejectsMalformedManifest(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := os.WriteFile(filepath.Join(repo, "open.riido.json"), []byte("{}{}"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = run(options{Repo: repo, Manifest: "open.riido.json"})
	if err == nil || !strings.Contains(err.Error(), "manifest must contain one JSON object") {
		t.Fatalf("expected malformed manifest error, got %v", err)
	}
}
