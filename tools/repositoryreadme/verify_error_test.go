package main

import (
	"path/filepath"
	"testing"
)

func TestVerifyShapeRejectsMissingIdentityFieldsAndContent(t *testing.T) {
	m := minimalReadmeManifest()
	m.SchemaVersion = "wrong"
	if err := verifyShape(m); err == nil {
		t.Fatal("expected identity error")
	}
	m = minimalReadmeManifest()
	m.Title = ""
	if err := verifyShape(m); err == nil {
		t.Fatal("expected required field error")
	}
	m = minimalReadmeManifest()
	m.Summary = nil
	if err := verifyShape(m); err == nil {
		t.Fatal("expected content error")
	}
	m = minimalReadmeManifest()
	m.Loop.Execute = ""
	if err := verifyShape(m); err == nil {
		t.Fatal("expected loop error")
	}
}

func TestVerifyAllRejectsMissingWorkflowAndDocLink(t *testing.T) {
	m := minimalReadmeManifest()
	root, _ := newReadmeRepo(t, m)
	if err := verifyAll(root, m, renderDoc(m)); err != nil {
		t.Fatal(err)
	}
	if err := verifyAll(t.TempDir(), m, renderDoc(m)); err == nil {
		t.Fatal("expected missing workflow error")
	}
	missingDocRoot := t.TempDir()
	writeReadmeTestFile(t, filepath.Join(missingDocRoot, m.Workflow), "name: readme\n")
	if err := verifyAll(missingDocRoot, m, renderDoc(m)); err == nil {
		t.Fatal("expected missing doc link error")
	}
}

func TestVerifyRenderedRejectsMissingAndForbiddenMarkers(t *testing.T) {
	m := minimalReadmeManifest()
	if err := verifyRendered("missing", m); err == nil {
		t.Fatal("expected missing required marker error")
	}
	m.ForbiddenLiterals = []string{"bad"}
	if err := verifyRendered(renderDoc(m)+"bad", m); err == nil {
		t.Fatal("expected forbidden literal error")
	}
}
