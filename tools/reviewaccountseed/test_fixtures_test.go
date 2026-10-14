package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeReviewSeedTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeReviewSeedManifest(t *testing.T, path string, m manifest) {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeReviewSeedTestFile(t, path, string(body))
}

func minimalReviewSeedManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-review-account-seed-evidence.v1",
		ID:               "review-seed-test",
		Title:            "Review Seed Test",
		GeneratedDoc:     "docs/generated.md",
		Workflow:         ".github/workflows/review-account-seed.yml",
		EvidenceArtifact: "review-account-seed-evidence",
		SeedSSOT:         "seed.riido.json",
		OwnerPackages:    []string{"internal/riidoaiserver"},
		Loop: evidenceLoop{
			Observation:   "observation",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retrospective",
		},
	}
}

func newReviewSeedRepo(t *testing.T, m manifest) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeReviewSeedTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	writeReviewSeedTestFile(t, filepath.Join(root, m.SeedSSOT), "redacted seed\n")
	manifestPath := filepath.Join(root, "manifest.json")
	writeReviewSeedManifest(t, manifestPath, m)
	return root, manifestPath
}
