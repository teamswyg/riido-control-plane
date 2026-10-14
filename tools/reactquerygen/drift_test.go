package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateReactQueryClientDoesNotDrift(t *testing.T) {
	spec := loadTestOpenAPI(t)
	got, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	want, err := os.ReadFile(testGeneratedClientPath())
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated React Query client drifted; run go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts")
	}
	gotReact, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	wantReact, err := os.ReadFile(testGeneratedReactClientPath())
	if err != nil {
		t.Fatalf("read generated react file: %v", err)
	}
	if !bytes.Equal(gotReact, wantReact) {
		t.Fatalf("generated React Query hook wrapper drifted; run go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts")
	}
}

func testGeneratedClientPath() string {
	return filepath.Join("..", "..", "web", "generated", "aiAgentClient.ts")
}

func testGeneratedReactClientPath() string {
	return filepath.Join("..", "..", "web", "generated", "aiAgentClient.react.ts")
}
