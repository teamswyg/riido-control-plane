package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadLiveSummaryCarriesSafeClaims(t *testing.T) {
	t.Setenv("TESTNET_TOKEN", "super-secret-token")
	t.Setenv("TESTNET_BASE_URL", "https://private.example.test")
	out := filepath.Join(t.TempDir(), "summary.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-workflow", "ai-agent-client-testnet-load",
		"-live-status", "failure",
		"-evidence-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	assertSummaryOmitsSecrets(t, text)
	for _, required := range loadSummaryClaimPhrases() {
		if !strings.Contains(text, required) {
			t.Fatalf("summary missing %s: %s", required, text)
		}
	}
}

func loadSummaryClaimPhrases() []string {
	return []string{
		"\"evidence_claims\"",
		"\"load_harness_client_read_pressure\"",
		"\"load_harness_closed_loop_promotion\"",
		"\"status\": \"not_verified\"",
	}
}
