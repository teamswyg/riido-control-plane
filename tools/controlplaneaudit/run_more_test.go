package main

import (
	"path/filepath"
	"testing"
)

func TestRunRejectsMissingRepoRoot(t *testing.T) {
	t.Parallel()
	err := run(options{Repo: filepath.Join(t.TempDir(), "missing")})
	assertErrorContains(t, err, "repository root not found")
}

func TestRunPropagatesEvidenceWriteFailure(t *testing.T) {
	t.Parallel()
	err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		EvidenceOut: t.TempDir(),
	})
	assertErrorContains(t, err, "is a directory")
}

func TestMaybeDocPropagatesMissingDocRead(t *testing.T) {
	t.Parallel()
	err := maybeDoc(t.TempDir(), manifest{GeneratedDoc: "missing.md"}, "doc", false, true)
	assertErrorContains(t, err, "missing.md")
}

func TestNewEvidencePropagatesSurfaceScanFailure(t *testing.T) {
	t.Parallel()
	_, err := newEvidence(t.TempDir(), manifest{
		Surfaces: []surface{{ID: "x", Files: []string{"missing.go"}}},
	})
	assertErrorContains(t, err, "missing.go")
}
