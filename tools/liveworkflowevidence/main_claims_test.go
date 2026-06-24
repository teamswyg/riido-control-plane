package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSmokeLiveSummaryCarriesSafeClaims(t *testing.T) {
	t.Setenv("TESTNET_TOKEN", "super-secret-token")
	t.Setenv("TESTNET_BASE_URL", "https://private.example.test")
	out := filepath.Join(t.TempDir(), "summary.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-workflow", "ai-agent-client-testnet-smoke",
		"-live-status", "success",
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
	for _, required := range smokeSummaryClaimPhrases() {
		if !strings.Contains(text, required) {
			t.Fatalf("summary missing %s: %s", required, text)
		}
	}
}

func assertSummaryOmitsSecrets(t *testing.T, text string) {
	t.Helper()
	for _, forbidden := range []string{"super-secret-token", "private.example.test"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("summary leaked %s: %s", forbidden, text)
		}
	}
}

func smokeSummaryClaimPhrases() []string {
	return []string{
		"\"evidence_claims\"",
		"\"v2_threads_active_stream_href\"",
		"\"v3_history_message_retention\"",
		"\"status\": \"verified\"",
	}
}
