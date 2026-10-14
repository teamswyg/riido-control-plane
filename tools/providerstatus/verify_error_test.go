package main

import (
	"path/filepath"
	"testing"
)

func TestProviderStatusVerifyHeaderRejectsIdentityAndRequiredFields(t *testing.T) {
	if err := verifyHeader(manifest{}); err == nil {
		t.Fatal("expected identity error")
	}
	m := minimalProviderStatusManifest()
	m.Title = ""
	if err := verifyHeader(m); err == nil {
		t.Fatal("expected required field error")
	}
}

func TestProviderStatusVerifySourcesRejectsEmptyMissingAndNeedleDrift(t *testing.T) {
	root := t.TempDir()
	if err := verifySources(root, nil); err == nil {
		t.Fatal("expected missing source checks error")
	}
	checks := []sourceCheck{{Name: "missing", File: "missing.go"}}
	if err := verifySources(root, checks); err == nil {
		t.Fatal("expected missing file error")
	}
	writeProviderStatusTestFile(t, filepath.Join(root, "source.go"), "needle")
	checks = []sourceCheck{{Name: "needle", File: "source.go", Contains: []string{"absent"}}}
	if err := verifySources(root, checks); err == nil {
		t.Fatal("expected missing needle error")
	}
}

func TestProviderStatusVerifyDocRejectsMissingAndStaleGeneratedDoc(t *testing.T) {
	root := t.TempDir()
	m := completeProviderStatusManifest()
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected missing doc error")
	}
	writeProviderStatusTestFile(t, filepath.Join(root, m.GeneratedDoc), "stale")
	if err := verifyDoc(root, m); err == nil {
		t.Fatal("expected stale doc error")
	}
}

func TestProviderStatusVerifyAcceptsMatchingManifestWithoutDocCheck(t *testing.T) {
	root := t.TempDir()
	writeProviderStatusTestFile(t, filepath.Join(root, "source.go"), "needle")
	m := completeProviderStatusManifest()
	m.SourceChecks = []sourceCheck{{Name: "source", File: "source.go", Contains: []string{"needle"}}}
	if err := verify(root, m, false); err != nil {
		t.Fatal(err)
	}
}
