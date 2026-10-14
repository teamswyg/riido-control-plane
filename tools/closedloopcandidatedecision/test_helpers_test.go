package main

import "testing"

func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func pinCandidateFreshnessClock(t *testing.T) {
	t.Helper()
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T01:00:00Z")
}
