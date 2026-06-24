package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-check-doc", "-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, phrase := range []string{"\"status\": \"verified\"", "\"workflow_count\": 3"} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("evidence missing %s: %s", phrase, text)
		}
	}
}

func TestLiveSummaryRedactsRuntimeValues(t *testing.T) {
	t.Setenv("TESTNET_TOKEN", "super-secret-token")
	t.Setenv("TESTNET_BASE_URL", "https://private.example.test")
	t.Setenv("RIIDO_LIVE_WORKFLOW_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	out := filepath.Join(t.TempDir(), "summary.json")
	err := mainRun([]string{
		"-repo", "../..",
		"-workflow", "deploy-ai-agent-testnet",
		"-deployment-mode", "ecs-rolling",
		"-build-cache-mode", "buildkit-gha",
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
	for _, forbidden := range []string{"super-secret-token", "private.example.test"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("summary leaked %s: %s", forbidden, text)
		}
	}
	for _, required := range []string{
		"\"generated_at\": \"2026-06-24T00:00:00Z\"",
		"\"expires_at\": \"2026-06-25T00:00:00Z\"",
		"\"deployment_mode\": \"ecs-rolling\"",
		"\"build_cache_mode\": \"buildkit-gha\"",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("summary missing %s: %s", required, text)
		}
	}
}
