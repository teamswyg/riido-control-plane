package main

import (
	"strings"
	"testing"
)

func TestRunRejectsFlagAndLoadErrors(t *testing.T) {
	t.Parallel()
	if err := run([]string{"-bad"}); err == nil {
		t.Fatal("expected flag parse error")
	}
	err := run([]string{"-contract", t.TempDir() + "/missing.json"})
	if err == nil || !strings.Contains(err.Error(), "read contract") {
		t.Fatalf("expected load error, got %v", err)
	}
}

func TestRunRejectsModuleListingFailure(t *testing.T) {
	t.Parallel()
	contractPath := writeContract(t, `{
		"schema_version": "riido-go-dependency-allowlist.v2",
		"id": "test-dependency-allowlist",
		"service": "riido-control-plane",
		"policy": "test",
		"assertions": ["direct modules must be approved"],
		"loop": {
			"observation": "test",
			"hypothesis": "test",
			"execute": "test",
			"evaluate": "test",
			"retrospective": "test"
		},
		"allowed_direct_modules": []
	}`)
	err := run([]string{"-contract", contractPath, "-module", t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "go list -m -json all") {
		t.Fatalf("expected module listing error, got %v", err)
	}
}
