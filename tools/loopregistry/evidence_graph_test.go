package main

import "testing"

func TestLoopRegistryEvidenceExposesGraphEdges(t *testing.T) {
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
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		t.Fatalf("verifyAll: %v", err)
	}
	got := newEvidence(m, result, nil)
	if len(got.EvidenceGraph) != len(m.EvidenceGraph) {
		t.Fatalf("evidence graph edges = %d, want %d", len(got.EvidenceGraph), len(m.EvidenceGraph))
	}
	for _, edge := range got.EvidenceGraph {
		if edge.From == "" || edge.To == "" || edge.Relation == "" {
			t.Fatalf("incomplete evidence graph edge: %+v", edge)
		}
	}
}
