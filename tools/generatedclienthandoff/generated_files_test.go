package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedClientHandoffWritesManifestHistoryReadmeAndPRBody(t *testing.T) {
	out := t.TempDir()
	prBodyPath := filepath.Join(out, "PR_BODY.generated.md")
	cfg := baseTestConfig(out)
	cfg.PRBody = prBodyPath
	if err := run(cfg); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertGeneratedClientHandoffFiles(t, out)
	assertGeneratedClientHandoffHashes(t, out)
	assertFileContains(t, filepath.Join(out, "contractManifest.generated.ts"), "v2.aiAgent.tasks.assign")
	assertFileContains(t, filepath.Join(out, "contractManifest.generated.ts"), "deprecated:")
	assertFileContains(t, filepath.Join(out, "apiHistory.generated.ts"), "assignment_ready")
	assertFileContains(t, filepath.Join(out, "README.generated.md"), "Lifecycle")
	assertFileContains(t, filepath.Join(out, "README.generated.md"), "X-Riido-AI-Agent-Token")
	assertFileContains(t, prBodyPath, "변경 요약")
	assertFileContains(t, prBodyPath, "SSOT 기반 결정사항")
	assertFileContains(t, prBodyPath, "team_id")
	assertFileContains(t, prBodyPath, "pnpm run type-check")
}

func TestGeneratedClientHandoffRejectsNonRiidoWorkBranch(t *testing.T) {
	out := t.TempDir()
	cfg := baseTestConfig(out)
	cfg.TargetBranch = "react-query-v0.0.99-0123456"
	err := run(cfg)
	assertErrorContains(t, err, "Riido work branchName")
}

func assertGeneratedClientHandoffHashes(t *testing.T, out string) {
	t.Helper()
	want := map[string]string{
		"README.generated.md":           "d064a582fe2ac10dca8d7ec4eb065c48328196f3cb4b8840394ecca37321fe7c",
		"PR_BODY.generated.md":          "0f9c21f8232e0015257ebc0b993a80ee63587465d124b3c694fe33a702896f60",
		"apiHistory.generated.ts":       "24abcf74d266c9c0c56b847d3f51aef9940217ae46565e39c2bc0c890ac66232",
		"contractManifest.generated.ts": "52a61d3d915d83e921a3c7ffa5a0dd1f5851d8136d6423a4a12e1a4f5accd5d7",
		"index.ts":                      "4f8d1b47853ebfa45c713d0ee5e40c13d19bfcb136418e6b2f4769f494909508",
		"react.ts":                      "079eead5d4755055e92e09c456893f45fdc10eadbbbb1f32bd1b26ac893136c3",
	}
	for name, hash := range want {
		data, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatalf("read generated file %s: %v", name, err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != hash {
			t.Fatalf("%s hash = %s, want %s", name, got, hash)
		}
	}
}

func assertGeneratedClientHandoffFiles(t *testing.T, out string) {
	t.Helper()
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
}
