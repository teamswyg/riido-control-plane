package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func baseTestConfig(out string) config {
	root := filepath.Join("..", "..")
	return config{
		OpenAPI:      filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json"),
		DSL:          filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.dsl.riido.json"),
		IR:           filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.ir.riido.json"),
		Core:         filepath.Join(root, "web", "generated", "aiAgentClient.ts"),
		React:        filepath.Join(root, "web", "generated", "aiAgentClient.react.ts"),
		Out:          out,
		SourceCommit: "0123456789abcdef",
		SourceRef:    "v0.0.99",
		TargetBranch: "A-60-AI-Agent-generated-client-handoff-test",
		GeneratedAt:  "2026-06-03",
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s missing %q", path, want)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}
