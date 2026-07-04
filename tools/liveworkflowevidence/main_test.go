package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	liveManifestEvidenceSHA256 = "a4ccf7c90c5e98aea521369c276fb5776364fb164edd01214a0cf1bd04b2d9df"
	liveDeploySummarySHA256    = "4816fc4fec1c718a5d6bf8a4b0b3fcc1e49f184fb040ea7f040a611b61d3d84e"
)

func TestLiveWorkflowEvidenceBehaviorGolden(t *testing.T) {
	t.Run("manifest", TestManifestEvidence)
	t.Run("summary", TestLiveSummaryRedactsRuntimeValues)
}

func TestManifestEvidence(t *testing.T) {
	t.Setenv("RIIDO_LIVE_WORKFLOW_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	manifestOut := filepath.Join(t.TempDir(), "manifest.json")
	if err := mainRun([]string{"-repo", "../..", "-check-doc", "-evidence-out", manifestOut}); err != nil {
		t.Fatal(err)
	}
	assertSHA256File(t, manifestOut, liveManifestEvidenceSHA256)
}

func TestLiveSummaryRedactsRuntimeValues(t *testing.T) {
	t.Setenv("RIIDO_LIVE_WORKFLOW_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	t.Setenv("TESTNET_TOKEN", "super-secret-token")
	t.Setenv("TESTNET_BASE_URL", "https://private.example.test")
	summaryOut := filepath.Join(t.TempDir(), "summary.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-workflow", "deploy-ai-agent-testnet",
		"-deployment-mode", "ecs-rolling",
		"-build-cache-mode", "buildkit-gha",
		"-evidence-out", summaryOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	data := assertSHA256File(t, summaryOut, liveDeploySummarySHA256)
	for _, forbidden := range []string{"super-secret-token", "private.example.test"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("summary leaked %s: %s", forbidden, data)
		}
	}
}

func assertSHA256File(t *testing.T, path, want string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("%s sha256 = %s, want %s\n%s", path, got, want, data)
	}
	return data
}
