package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, repo, slashPath, body string) {
	t.Helper()
	path := filepath.Join(repo, filepath.FromSlash(slashPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runtimeBoundaryTestRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	writeFile(t, repo, "go.mod", "module example.com/runtimeboundary\n")
	writeFile(t, repo, "runtime-cd.json", runtimeBoundaryLinkedManifest())
	writeFile(t, repo, "evidence.txt", "runtime deploy boundary evidence\n")
	return repo
}

func runtimeBoundaryLinkedManifest() string {
	return `{
  "schema_version": "riido-control-plane-runtime-cd-ownership.v1",
  "id": "runtime-cd-ownership",
  "current_strategy": {"workflow": ".github/workflows/deploy-ai-agent-testnet.yml"}
}`
}

func runtimeBoundaryManifest(generatedDoc string) manifest {
	return manifest{
		SchemaVersion: manifestSchema,
		ID:            "runtime-boundary-test",
		Title:         "Runtime Boundary Test",
		GeneratedDoc:  generatedDoc,
		Workflow:      ".github/workflows/runtime-deployment-boundary.yml",
		Evidence:      "runtime-boundary-evidence",
		LinkedCD:      "runtime-cd.json",
		Boundaries: []boundary{{
			ID:       "deploy",
			Owner:    "control-plane",
			Scope:    "runtime deployment",
			Evidence: []evidenceCheck{{Path: "evidence.txt", Contains: []string{"runtime deploy"}}},
		}},
		Rules: []string{"runtime deployments must be workflow-owned"},
		Loop:  loopRecord{"observe", "hypothesis", "execute", "evaluate", "retrospective"},
	}
}
