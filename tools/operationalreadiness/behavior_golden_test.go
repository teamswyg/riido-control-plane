package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestOperationalReadinessBehaviorGolden(t *testing.T) {
	for key, value := range map[string]string{
		readinessNowEnv:      "2026-07-04T00:00:00Z",
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
	assertFileSHA256(t, evidenceOut, "693bc7152851e3cee00303302c17e87e5944677c9db5679c3e82f163e37123e7")
	assertFileSHA256(t, candidateOut, "87e7b26e1bfa9630ce8790734c5e5fb9af80c02d98eb7d0b36daf85c15e79644")
	assertFileSHA256(t, publicOut, "16ba6247b71b5369ae87fc79c484c19d933851de4b43d72d8eb9faeaa16eaf2b")
	e := readEvidence(t, evidenceOut)
	if e.CheckCount != 14 || e.MeasurementCount != 89 || e.PartialCount != 9 ||
		e.StalePartialCount != 9 || e.PartialPromotion.CandidateCount != 9 ||
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
