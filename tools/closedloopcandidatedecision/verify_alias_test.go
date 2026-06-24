package main

import "testing"

func TestCandidateDecisionManifestVerifies(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(root, out); err != nil {
		t.Fatal(err)
	}
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CandidateIn: out,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestCandidateDecisionVerifyAlias(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(root, out); err != nil {
		t.Fatal(err)
	}
	if err := mainRun([]string{
		"-repo", "../..", "-manifest", defaultManifest,
		"-candidate-in", out, "-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}
