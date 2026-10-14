package main

import "testing"

func TestEvidenceGraphRejectsUnclaimedChain(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	m.Chains[0].Claims = nil
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected unclaimed evidence chain to fail")
	}
}
