package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMainRunParsesFlagsAndRuns(t *testing.T) {
	repo := openQuestionsTempRepo(t, validOpenQuestionsManifest())
	evidenceOut := filepath.Join(repo, "out", "evidence.json")

	err := mainRun([]string{
		"-repo", repo,
		"-manifest", "open.riido.json",
		"-write-doc",
		"-check-doc",
		"-evidence-out", evidenceOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(repo, "docs", "open.md"), generatedNotice)
	assertFileContains(t, evidenceOut, "verified")
}

func TestMainRunRejectsBadFlag(t *testing.T) {
	err := mainRun([]string{"-unknown"})
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("expected flag error, got %v", err)
	}
}
