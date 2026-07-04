package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

const closedLoopCandidateDecisionGoldenSHA256 = "78d198b1a349193d59c9669f8040139586967f6541477cd97fe0828434b60af9"

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
