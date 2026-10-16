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
		"README.generated.md":           "783fed97037e711f2e7d5fbc044c36c7ecef785099e598e2fc82b786c386d573",
		"PR_BODY.generated.md":          "40ff31940ce643a87715cbe926c25dcc5e0649039a141924cae0e67e7ff28c46",
		"apiHistory.generated.ts":       "d1b692d025ed0dc6e4c138f0de0db52506262cc97acb30c4ff00d8784b894c3c",
		"contractManifest.generated.ts": "e65191e43aa763e552ef532460f5b9dd2b535b2bbe34081a0236d4298cdea8f0",
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
