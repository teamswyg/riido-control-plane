package main

import "testing"

func TestLoopEvidenceKindsRejectUnknownKind(t *testing.T) {
	err := verifyEvidenceKinds(manifest{
		EvidenceKinds: []evidenceKind{{Kind: "json", Description: "structured json"}},
		Loops: []loopRecord{{
			ID:       "loop-a",
			Evidence: []evidenceSource{{Kind: "screenshot", Path: "screen.png"}},
		}},
	})
	if err == nil {
		t.Fatal("expected unknown evidence kind to fail")
	}
}

func TestLoopEvidenceKindsRejectMissingVocabulary(t *testing.T) {
	err := verifyEvidenceKinds(manifest{
		Loops: []loopRecord{{ID: "loop-a", Evidence: []evidenceSource{{Kind: "json"}}}},
	})
	if err == nil {
		t.Fatal("expected missing evidence kind vocabulary to fail")
	}
}

func TestLoopEvidenceKindsAcceptManifest(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	if err := verifyEvidenceKinds(m); err != nil {
		t.Fatal(err)
	}
}
