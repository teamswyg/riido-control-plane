package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeReadmeTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeReadmeManifest(t *testing.T, path string, m manifest) {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeReadmeTestFile(t, path, string(body))
}

func minimalReadmeManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		ID:               "readme-test",
		Title:            "README Test",
		RiidoTask:        "RIID-test",
		GeneratedDoc:     generatedDoc,
		Workflow:         ".github/workflows/repository-readme.yml",
		EvidenceArtifact: "repository-readme-evidence",
		Summary:          []string{"summary marker"},
		DocLinks:         []docLink{{Topic: "doc", Path: "docs/doc.md"}},
		Verification:     []string{"go test ./tools/repositoryreadme"},
		RequiredMarkers: []string{
			"summary marker",
		},
		Loop: evidenceLoop{
			Observation:   "observation",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retrospective",
		},
	}
}

func newReadmeRepo(t *testing.T, m manifest) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeReadmeTestFile(t, filepath.Join(root, m.Workflow), "name: readme\n")
	writeReadmeTestFile(t, filepath.Join(root, "docs/doc.md"), "doc\n")
	manifestPath := filepath.Join(root, "README.riido.json")
	writeReadmeManifest(t, manifestPath, m)
	return root, manifestPath
}
