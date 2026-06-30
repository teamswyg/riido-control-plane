package main

import "testing"

func TestHarnessPromotionEvidenceCarriesExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll("../..", m)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if got.GeneratedAt != "2026-06-24T00:00:00Z" {
		t.Fatalf("generated_at = %q", got.GeneratedAt)
	}
	if got.ExpiresAt != "2026-06-25T00:00:00Z" {
		t.Fatalf("expires_at = %q", got.ExpiresAt)
	}
}
