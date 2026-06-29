package main

import "testing"

func TestCandidateIntakePreservesGeneratedHarnessSubject(t *testing.T) {
	root := repoRootForTest(t)
	out := candidateFixtureForTest(t, root)
	result, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CandidateSubjects) != 1 {
		t.Fatalf("candidate subjects = %+v", result.CandidateSubjects)
	}
	subject := result.CandidateSubjects[0]
	if subject.Kind != "harness_failure" ||
		subject.CandidateID != "ai-agent-client-testnet-smoke:provider_smoke" {
		t.Fatalf("candidate subject = %+v", subject)
	}
}
