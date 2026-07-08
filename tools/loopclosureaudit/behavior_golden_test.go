package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

const loopClosureAuditGoldenSHA256 = "3cd675ce59910c23ed654ad89a5e281c59c819c4922f868bee3f507e6fc80851"

func TestLoopClosureAuditBehaviorGolden(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T12:00:00Z")
	t.Setenv("RIIDO_LOOP_CLOSURE_AUDIT_NOW", "2026-06-24T12:00:00Z")
	dir := t.TempDir()
	evidenceOut := dir + "/evidence.json"
	candidateOut := dir + "/candidates.json"
	err := run(options{
		Repo: "../..", Manifest: defaultManifest,
		CheckDoc: true, EvidenceOut: evidenceOut, CandidateOut: candidateOut,
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
	body = normalizeSemanticHashes(body)
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(body))); got != loopClosureAuditGoldenSHA256 {
		t.Fatalf("loop closure audit SHA mismatch: got %s want %s\n%s", got, loopClosureAuditGoldenSHA256, body)
	}
}

func normalizeSemanticHashes(body string) string {
	re := regexp.MustCompile(`"semantic_hash": "[0-9a-f]{64}"`)
	return re.ReplaceAllString(body, `"semantic_hash": "<semantic-hash>"`)
}
