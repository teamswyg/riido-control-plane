package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadContractRejectsUnapprovedEntry(t *testing.T) {
	path := writeContract(t, `{
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
		"allowed_direct_modules": [
			{"path": "go.opentelemetry.io/otel", "layer": "observability", "owner": "platform", "approved": false, "reason": "test"}
		]
	}`)
	_, err := loadContract(path)
	if err == nil || !strings.Contains(err.Error(), "approved must be true") {
		t.Fatalf("expected approved flag error, got %v", err)
	}
}

func TestLoadContractRejectsUnknownLayer(t *testing.T) {
	path := writeContract(t, `{
		"schema_version": "riido-go-dependency-allowlist.v2",
		"id": "test-dependency-allowlist",
		"service": "riido-control-plane",
		"policy": "test",
		"assertions": ["direct modules must use a known layer"],
		"loop": {
			"observation": "test",
			"hypothesis": "test",
			"execute": "test",
			"evaluate": "test",
			"retrospective": "test"
		},
		"allowed_direct_modules": [
			{"path": "github.com/example/config-framework", "layer": "framework", "owner": "platform", "approved": true, "reason": "test"}
		]
	}`)
	_, err := loadContract(path)
	if err == nil || !strings.Contains(err.Error(), "not in vocabulary") {
		t.Fatalf("expected layer vocabulary error, got %v", err)
	}
}

func writeContract(t *testing.T, data string) string {
	t.Helper()
	path := t.TempDir() + "/dependency_allowlist.riido.json"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}
