package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

const harnessPromotionGoldenSHA256 = "8ed52496e86e2bb3dbb0c609dc5dea17d5de14cde7dfd1ccd095079cc005af1a"

func TestHarnessPromotionBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T01:00:00Z")
	t.Setenv("RIIDO_HARNESS_PROMOTION_NOW", "2026-06-24T01:00:00Z")
	dir := t.TempDir()
	candidateOut := dir + "/candidates.json"
	evidenceOut := dir + "/evidence.json"
	err := run(options{
		Repo: "../..", Manifest: defaultManifest,
		Summary:      "docs/30-architecture/fixtures/harness-failure-summary.fixture.json",
		CandidateOut: candidateOut, EvidenceOut: evidenceOut, CheckDoc: true,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	evidenceBody, err := os.ReadFile(evidenceOut)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	candidateBody, err := os.ReadFile(candidateOut)
	if err != nil {
		t.Fatalf("read candidate: %v", err)
	}
	body := string(evidenceBody) + "\n---candidate---\n" + string(candidateBody)
	body = strings.ReplaceAll(body, candidateOut, "<candidate-fixture>")
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(body))); got != harnessPromotionGoldenSHA256 {
		t.Fatalf("harness promotion SHA mismatch: got %s want %s\n%s", got, harnessPromotionGoldenSHA256, body)
	}
}
