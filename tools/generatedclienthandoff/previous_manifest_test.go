package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedClientHandoffPRBodyIncludesPreviousManifestDiff(t *testing.T) {
	out := t.TempDir()
	previousManifestPath := filepath.Join(out, "previous-contractManifest.generated.ts")
	prBodyPath := filepath.Join(out, "PR_BODY.generated.md")
	if err := os.WriteFile(previousManifestPath, []byte(previousManifestFixture), 0o644); err != nil {
		t.Fatalf("write previous manifest: %v", err)
	}
	cfg := baseTestConfig(out)
	cfg.PRBody = prBodyPath
	cfg.PreviousManifest = previousManifestPath
	if err := run(cfg); err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, want := range []string{"이전 operation 수", "이전 source commit: `previous-source-commit`", "추가된 generated paths", "제거된 generated paths", "`aiAgent.removed.example`", "변경된 generated paths", "`aiAgent.bootstrap`", "getAIAgentClientBootstrapOld"} {
		assertFileContains(t, prBodyPath, want)
	}
}

const previousManifestFixture = `export const riidoControlPlaneContractManifest = {
  sourceCommit: 'previous-source-commit',
  operations: [
    { generatedPath: 'aiAgent.bootstrap', operationId: 'getAIAgentClientBootstrapOld', method: 'GET', path: '/v1/client/ai-agent/bootstrap-old', deprecated: false },
    { generatedPath: 'aiAgent.removed.example', operationId: 'removedExample', method: 'DELETE', path: '/v1/client/ai-agent/removed-example', deprecated: false },
  ],
} as const;
`

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
