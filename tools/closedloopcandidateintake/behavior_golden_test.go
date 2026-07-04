package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

const closedLoopCandidateIntakeGoldenSHA256 = "3f8c86d88cc8887bb54f4e52f0d1e5426fce7ee9d9e26c98feaf614b9cc46f62"

func TestClosedLoopCandidateIntakeBehaviorGolden(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
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
	if got := fmt.Sprintf("%x", sha256.Sum256(body)); got != closedLoopCandidateIntakeGoldenSHA256 {
		t.Fatalf("candidate intake SHA mismatch: got %s want %s\n%s", got, closedLoopCandidateIntakeGoldenSHA256, body)
	}
}
