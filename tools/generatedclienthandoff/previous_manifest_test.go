package main

import (
	"os"
	"path/filepath"
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
	assertFileContains(t, prBodyPath, "이전 operation 수")
	assertFileContains(t, prBodyPath, "이전 source commit: `previous-source-commit`")
	assertFileContains(t, prBodyPath, "추가된 generated paths")
	assertFileContains(t, prBodyPath, "제거된 generated paths")
	assertFileContains(t, prBodyPath, "`aiAgent.removed.example`")
	assertFileContains(t, prBodyPath, "변경된 generated paths")
	assertFileContains(t, prBodyPath, "`aiAgent.bootstrap`")
	assertFileContains(t, prBodyPath, "getAIAgentClientBootstrapOld")
}

const previousManifestFixture = `export const riidoControlPlaneContractManifest = {
  sourceCommit: 'previous-source-commit',
  operations: [
    { generatedPath: 'aiAgent.bootstrap', operationId: 'getAIAgentClientBootstrapOld', method: 'GET', path: '/v1/client/ai-agent/bootstrap-old', deprecated: false },
    { generatedPath: 'aiAgent.removed.example', operationId: 'removedExample', method: 'DELETE', path: '/v1/client/ai-agent/removed-example', deprecated: false },
  ],
} as const;
`
