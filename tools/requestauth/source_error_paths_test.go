package main

import (
	"strings"
	"testing"
)

func TestVerifySourcesRequiresAtLeastOneSourceCheck(t *testing.T) {
	t.Parallel()
	if err := verifySources(t.TempDir(), nil); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("verifySources error = %v, want required", err)
	}
}

func TestVerifySourcesRejectsMissingFileAndNeedle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	missing := []sourceCheck{{Name: "missing", File: "missing.go", Contains: []string{"needle"}}}
	if err := verifySources(root, missing); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("verifySources missing file error = %v", err)
	}
	writeFile(t, root, "source.txt", "haystack")
	wrongNeedle := []sourceCheck{{Name: "source", File: "source.txt", Contains: []string{"needle"}}}
	if err := verifySources(root, wrongNeedle); err == nil || !strings.Contains(err.Error(), "needle") {
		t.Fatalf("verifySources needle error = %v", err)
	}
}

func TestVerifyRunsAllSections(t *testing.T) {
	t.Parallel()
	root := writeTestRepo(t, testManifest(), "needle")
	m, err := loadManifest(resolve(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := verify(root, m, false); err != nil {
		t.Fatal(err)
	}
}
