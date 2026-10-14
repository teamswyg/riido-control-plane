package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeHealthReadyTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeHealthReadyTestManifest(t *testing.T, path string, m manifest) {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeHealthReadyTestFile(t, path, string(body))
}

func minimalHealthReadyManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-health-ready-cmd.v1",
		ID:               "health-ready-test",
		Title:            "Health Ready Test",
		GeneratedDoc:     "docs/generated.md",
		Workflow:         ".github/workflows/health-ready-cmd.yml",
		EvidenceArtifact: "health-ready-cmd-evidence",
		OwnerPackages:    []string{"internal/riidoaiserver"},
		Endpoints: []endpointContract{
			{Name: "health", Method: "GET", Path: "/healthz", Status: "ok", HTTPStatus: 200},
		},
		Loop: evidenceLoop{
			Observation:   "observed",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retro",
		},
	}
}

func newHealthReadyRepo(t *testing.T, m manifest) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeHealthReadyTestFile(t, filepath.Join(root, "go.mod"), "module test\n")
	manifestPath := filepath.Join(root, "manifest.json")
	writeHealthReadyTestManifest(t, manifestPath, m)
	return root, manifestPath
}
