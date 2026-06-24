package main

import "testing"

func TestEvidenceGraphEvidenceExposesFullChainSurface(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result)
	if len(got.Chains) == 0 {
		t.Fatal("chains missing")
	}
	for _, c := range got.Chains {
		if c.ID == "" || c.Observation == "" || c.Hypothesis == "" ||
			c.Decision == "" || c.NextLoop == "" {
			t.Fatalf("chain evidence missing reasoning fields: %+v", c)
		}
		if len(c.Changes) == 0 || len(c.Verifiers) == 0 || len(c.Evidence) == 0 {
			t.Fatalf("chain evidence missing executable refs: %+v", c)
		}
	}
}
