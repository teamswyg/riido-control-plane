package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

const closedLoopCandidateDecisionGoldenSHA256 = "3058d753c88942935e3e7b99026a019febf04829cae374cc078eac221a846a07"

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
