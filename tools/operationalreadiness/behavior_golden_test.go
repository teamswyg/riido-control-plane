package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestOperationalReadinessBehaviorGolden(t *testing.T) {
	for key, value := range map[string]string{
		readinessNowEnv:      "2026-07-05T00:00:00Z",
		"GITHUB_WORKFLOW":    "operational-readiness",
		"GITHUB_SHA":         "syntaxhashgolden",
		"GITHUB_RUN_ID":      "846000",
		"GITHUB_RUN_ATTEMPT": "1",
		"GITHUB_REF_NAME":    "codex/syntax-hash-operationalreadiness",
		"GITHUB_EVENT_NAME":  "pull_request",
		"GITHUB_SERVER_URL":  "https://github.com",
		"GITHUB_REPOSITORY":  "teamswyg/riido-control-plane",
	} {
		t.Setenv(key, value)
	}
	dir := t.TempDir()
	evidenceOut := dir + "/evidence.json"
	candidateOut := dir + "/candidates.json"
	publicOut := dir + "/public.json"
	err := run(options{
		Repo: "../..", Manifest: defaultManifest, CheckDoc: true,
		EvidenceOut: evidenceOut, CandidateOut: candidateOut,
		PublicStatusJSON: publicOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileSHA256(t, evidenceOut, "2b20c121c22f0b72d8970f29db0b26696f80f0bf40151a0995990c6d5a8cf438")
	assertFileSHA256(t, candidateOut, "49dcfc4c93e974841f531f25606bd92219152411b6198e0684e8b80733afc55f")
	assertFileSHA256(t, publicOut, "808e195910331520054fe087c8ef79a35354b9eab0ddbd1d7ad49853077990cc")
	e := readEvidence(t, evidenceOut)
	if e.CheckCount != 14 || e.MeasurementCount != 89 || e.PartialCount != 9 ||
		e.StalePartialCount != 9 || e.PartialPromotion.CandidateCount != 9 ||
		e.Completion.InternalCompletenessBasisPoints != 10000 ||
		e.PublicStatus.P0CycleCount != 8 || e.PublicStatus.Overall != "degraded" ||
		e.PublicStatus.SourceWorkflow != "operational-readiness" ||
		e.PublicStatus.SourceCommit != "syntaxhashgolden" {
		t.Fatalf("readiness behavior drift: %+v", e)
	}
}

func assertFileSHA256(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("%s sha256 = %s", path, got)
	}
}
