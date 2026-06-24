package main

import "testing"

const candidateFixtureSummary = "docs/30-architecture/fixtures/harness-failure-summary.fixture.json"

func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func candidateFixtureForTest(t *testing.T, root string) string {
	t.Helper()
	out := t.TempDir() + "/candidates.json"
	if err := promoteSummary(root, candidateFixtureSummary, out); err != nil {
		t.Fatal(err)
	}
	return out
}
