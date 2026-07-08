package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyWorkflowFileRejectsMissingPhraseAndRawOutput(t *testing.T) {
	t.Parallel()
	m := liveTestManifest()
	root := writeLiveRepo(t, m, "go run ./tools/liveworkflowevidence\n")
	if _, err := verifyWorkflowFile(root, m.Workflows[0]); err == nil ||
		!strings.Contains(err.Error(), "missing phrase") {
		t.Fatalf("expected missing phrase error, got %v", err)
	}
	root = writeLiveRepo(t, m, validWorkflowText()+"cat out/private.json\n")
	if _, err := verifyWorkflowFile(root, m.Workflows[0]); err == nil ||
		!strings.Contains(err.Error(), "raw shell output") {
		t.Fatalf("expected raw output error, got %v", err)
	}
}

func TestRunRejectsUnknownWorkflowAndStaleDoc(t *testing.T) {
	t.Parallel()
	m := liveTestManifest()
	root := writeLiveRepo(t, m, validWorkflowText())
	if err := run(options{Repo: root, Manifest: filepath.Join(root, "manifest.json"), WorkflowID: "missing", EvidenceOut: filepath.Join(root, "out.json")}); err == nil {
		t.Fatal("expected unknown workflow error")
	}
	doc := repoPath(root, m.GeneratedDoc)
	if err := writeText(doc, "stale"); err != nil {
		t.Fatalf("write stale doc: %v", err)
	}
	if err := run(options{Repo: root, Manifest: filepath.Join(root, "manifest.json"), CheckDoc: true}); err == nil ||
		!strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}
