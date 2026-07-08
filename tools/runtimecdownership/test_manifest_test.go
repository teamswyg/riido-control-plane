package main

import "testing"

func testRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func testManifest(t *testing.T) manifest {
	t.Helper()
	root := testRepoRoot(t)
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
