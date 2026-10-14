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
	m, err := loadManifest(resolve(testRepoRoot(t), defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func testCaseByKind(t *testing.T, kind string) caseSpec {
	t.Helper()
	for _, tc := range testManifest(t).Cases {
		if tc.Kind == kind {
			return tc
		}
	}
	t.Fatalf("missing case kind %s", kind)
	return caseSpec{}
}
