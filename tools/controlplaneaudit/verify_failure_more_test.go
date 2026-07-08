package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifySurfacesRejectsShapeAndFileFailures(t *testing.T) {
	t.Parallel()
	for name, surfaces := range map[string][]surface{
		"missing-fields": {{ID: "x"}},
		"duplicate-id": {
			{ID: "x", Category: "endpoint", Risk: "risk", Candidate: "candidate", Files: []string{"go.mod"}, Patterns: []string{"module"}},
			{ID: "x", Category: "endpoint", Risk: "risk", Candidate: "candidate", Files: []string{"go.mod"}, Patterns: []string{"module"}},
		},
		"missing-files": {
			{ID: "x", Category: "endpoint", Risk: "risk", Candidate: "candidate", Patterns: []string{"module"}},
		},
		"missing-file": {
			{ID: "x", Category: "endpoint", Risk: "risk", Candidate: "candidate", Files: []string{"missing.go"}, Patterns: []string{"module"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := verifySurfaces("../..", surfaces); err == nil {
				t.Fatalf("verifySurfaces accepted %s", name)
			}
		})
	}
}

func TestVerifyWorkflowLoopAndPprofFailures(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workflow := filepath.Join(root, ".github/workflows/control-plane-performance.yml")
	if err := os.MkdirAll(filepath.Dir(workflow), 0o755); err != nil {
		t.Fatalf("mkdir workflow: %v", err)
	}
	if err := os.WriteFile(workflow, []byte("name: incomplete\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	m := validManifestForVerifyTest()
	m.Workflow = ".github/workflows/control-plane-performance.yml"
	if err := verifyWorkflow(root, m); err == nil {
		t.Fatal("verifyWorkflow accepted incomplete workflow")
	}
	if err := verifyLoop(loopSpec{Observation: "x"}); err == nil {
		t.Fatal("verifyLoop accepted incomplete loop")
	}
	if err := verifyPprofCommands([]string{"go tool pprof http://127.0.0.1:6060/debug/pprof/profile"}); err == nil {
		t.Fatal("verifyPprofCommands accepted incomplete pprof list")
	}
}
