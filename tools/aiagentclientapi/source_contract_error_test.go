package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/requirements"
)

func TestVerifySourcesRejectsEmptyMissingAndMissingNeedle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	assertAIClientAPIError(t, verifySources(root, nil), "source checks are required")
	check := sourceCheck{Name: "source", File: "source.go", Contains: []string{"needle"}}
	assertAIClientAPIError(t, verifySources(root, []sourceCheck{check}), "read source check")
	if err := writeText(filepath.Join(root, "source.go"), "haystack"); err != nil {
		t.Fatal(err)
	}
	assertAIClientAPIError(t, verifySources(root, []sourceCheck{check}), "missing \"needle\"")
}

func TestVerifyRequiredPathsRejectsOpenAPIAndSmokeGaps(t *testing.T) {
	t.Parallel()
	m := manifest{RequiredGeneratedPaths: []string{"custom.path"}}
	openapi := map[string]struct{}{}
	smoke := map[string]struct{}{}
	for _, path := range requirements.RequiredGeneratedPaths {
		openapi[path] = struct{}{}
		smoke[path] = struct{}{}
	}
	assertAIClientAPIError(t, verifyRequiredPaths(m, openapi, smoke), "missing from openapi")
	openapi["custom.path"] = struct{}{}
	assertAIClientAPIError(t, verifyRequiredPaths(m, openapi, smoke), "missing from smoke matrix")
}

func TestVerifyFileRejectsMissingAndAcceptsExisting(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "required.txt")
	assertAIClientAPIError(t, verifyFile(path), "required file")
	if err := writeText(path, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := verifyFile(path); err != nil {
		t.Fatalf("verifyFile existing file: %v", err)
	}
}
