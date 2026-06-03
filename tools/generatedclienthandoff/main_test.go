package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedClientHandoffWritesManifestHistoryReadmeAndPRBody(t *testing.T) {
	root := filepath.Join("..", "..")
	out := t.TempDir()
	prBodyPath := filepath.Join(out, "PR_BODY.generated.md")
	err := run(config{
		OpenAPI:      filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json"),
		DSL:          filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.dsl.riido.json"),
		IR:           filepath.Join(root, "contracts", "ai-agent-client", "control-plane-ai-agent-client.ir.riido.json"),
		Core:         filepath.Join(root, "web", "generated", "aiAgentClient.ts"),
		React:        filepath.Join(root, "web", "generated", "aiAgentClient.react.ts"),
		Out:          out,
		PRBody:       prBodyPath,
		SourceCommit: "0123456789abcdef",
		SourceRef:    "v0.0.99",
		TargetBranch: "react-query-v0.0.99-0123456",
		GeneratedAt:  "2026-06-03",
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, name := range []string{
		"README.generated.md",
		"apiHistory.generated.ts",
		"contractManifest.generated.ts",
		"index.ts",
		"react.ts",
	} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
	}
	assertFileContains(t, filepath.Join(out, "contractManifest.generated.ts"), "v2.aiAgent.tasks.assign")
	assertFileContains(t, filepath.Join(out, "contractManifest.generated.ts"), "deprecated:")
	assertFileContains(t, filepath.Join(out, "apiHistory.generated.ts"), "assignment_ready")
	assertFileContains(t, filepath.Join(out, "README.generated.md"), "Lifecycle")
	assertFileContains(t, filepath.Join(out, "README.generated.md"), "X-Riido-AI-Agent-Token")
	assertFileContains(t, prBodyPath, "SSOT 기반 결정사항")
	assertFileContains(t, prBodyPath, "team_id")
	assertFileContains(t, prBodyPath, "pnpm run type-check")
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
