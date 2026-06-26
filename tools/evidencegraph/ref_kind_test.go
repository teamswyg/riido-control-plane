package main

import "testing"

func TestEvidenceReferenceKindMustUseVocabulary(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	m.Chains[0].Changes[0].Kind = "invented_kind"

	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected unsupported reference kind to fail")
	}
}
