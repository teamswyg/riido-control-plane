package main

import "testing"

func loadLoopRegistryForTest(t *testing.T) (manifest, map[string]string) {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatal(err)
	}
	return m, hashes
}

func findLoopIndexForTest(t *testing.T, m manifest, id string) int {
	t.Helper()
	for i, loop := range m.Loops {
		if loop.ID == id {
			return i
		}
	}
	t.Fatalf("loop %s not found", id)
	return -1
}
