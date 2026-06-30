package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntentGateContractLiveSummaryCarriesSafeClaims(t *testing.T) {
	t.Setenv("RIIDO_LIVE_CHECK_STATUS", "success")
	out := filepath.Join(t.TempDir(), "summary.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-workflow", "ai-agent-intent-gate-contract",
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
	for _, required := range intentGateSummaryClaimPhrases() {
		if !strings.Contains(text, required) {
			t.Fatalf("summary missing %s: %s", required, text)
		}
	}
}

func intentGateSummaryClaimPhrases() []string {
	return []string{
		"\"evidence_claims\"",
		"\"intent_gate_waiting_for_user_no_stream_contract\"",
		"\"intent_gate_metadata_fallback_contract\"",
		"\"intent_gate_v3_conversation_timeline_contract\"",
		"\"intent_gate_closed_loop_promotion\"",
		"\"status\": \"verified\"",
	}
}
