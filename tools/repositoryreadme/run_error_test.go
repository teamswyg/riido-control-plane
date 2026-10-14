package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunParsesFlagsAndWritesEvidence(t *testing.T) {
	root, manifestPath := newReadmeRepo(t, minimalReadmeManifest())
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run([]string{"-repo", root, "-manifest", manifestPath, "-evidence-out", out})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsBadFlag(t *testing.T) {
	if err := run([]string{"-bad"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func TestRunWithOptionsRejectsWriteCheckAndEvidenceFailures(t *testing.T) {
	root, manifestPath := newReadmeRepo(t, minimalReadmeManifest())
	writeReadmeTestFile(t, filepath.Join(root, generatedDoc), "stale")
	if err := runWithOptions(root, manifestPath, false, true, ""); err == nil {
		t.Fatal("expected stale README error")
	}
	blocked := filepath.Join(root, "blocked")
	writeReadmeTestFile(t, blocked, "not a dir")
	err := runWithOptions(root, manifestPath, false, false, filepath.Join("blocked", "out.json"))
	if err == nil {
		t.Fatal("expected evidence write error")
	}
}

func TestRunWithOptionsWritesGeneratedDoc(t *testing.T) {
	root, manifestPath := newReadmeRepo(t, minimalReadmeManifest())
	if err := runWithOptions(root, manifestPath, true, true, ""); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(root, generatedDoc))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "summary marker") {
		t.Fatalf("generated README missing marker: %s", body)
	}
}
