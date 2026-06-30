package main

import (
	"os"
	"strings"
	"testing"
)

func TestTargetVerifierScriptPrefersFocusedCommands(t *testing.T) {
	body, err := targetVerifierScript(&targetVerifierPlan{
		FocusedCommands:    []string{"go test ./tools/loopregistry -count=1"},
		EntrypointCommands: []string{"go test ./..."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, "go test ./...") ||
		!strings.Contains(body, "go test ./tools/loopregistry -count=1") {
		t.Fatalf("unexpected script body: %s", body)
	}
}

func TestTargetVerifierScriptPrefersRunnableCommands(t *testing.T) {
	body, err := targetVerifierScript(&targetVerifierPlan{
		RunnableCommands:   []string{"go test ./tools/runnable -count=1"},
		FocusedCommands:    []string{"go test ./tools/focused -count=1"},
		EntrypointCommands: []string{"go test ./..."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, "go test ./tools/focused") ||
		strings.Contains(body, "go test ./...") ||
		!strings.Contains(body, "go test ./tools/runnable -count=1") {
		t.Fatalf("unexpected script body: %s", body)
	}
}

func TestWriteTargetVerifierScriptRequiresImpactPlan(t *testing.T) {
	err := writeTargetVerifierScript(t.TempDir()+"/verify.sh", &impactEvidence{})
	if err == nil {
		t.Fatal("expected missing impact plan to fail")
	}
}

func TestWriteTargetVerifierScriptCreatesExecutableFile(t *testing.T) {
	path := t.TempDir() + "/verify.sh"
	err := writeTargetVerifierScript(path, &impactEvidence{
		TargetVerifierPlan: &targetVerifierPlan{
			EntrypointCommands: []string{"go test ./tools/loopregistry -count=1"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("script mode = %v", info.Mode().Perm())
	}
}
