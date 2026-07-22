package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

const closedLoopCandidateDecisionGoldenSHA256 = "58f0577270e469fb1efc9c962d34ffff69388e0c99b464b943061a3d43d641d5"

func TestClosedLoopCandidateDecisionBehaviorGolden(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, candidateIn); err != nil {
		t.Fatalf("generate candidate: %v", err)
	}
	out := t.TempDir() + "/evidence.json"
	err := run(options{
		Repo: "../..", Manifest: defaultManifest,
		CandidateIn: candidateIn, CheckDoc: true, EvidenceOut: out,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	body = []byte(strings.ReplaceAll(string(body), candidateIn, "<candidate-fixture>"))
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != closedLoopCandidateDecisionGoldenSHA256 {
		t.Fatalf("candidate decision SHA mismatch: got %s want %s\n%s", got, closedLoopCandidateDecisionGoldenSHA256, body)
	}
}
