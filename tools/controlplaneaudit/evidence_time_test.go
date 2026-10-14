package main

import "testing"

func TestControlPlaneHighTrafficAuditEvidenceCarriesExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")

	got, err := newEvidence("../..", loadManifestForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	if got.GeneratedAt != "2026-06-24T00:00:00Z" {
		t.Fatalf("generated_at = %q", got.GeneratedAt)
	}
	if got.ExpiresAt != "2026-06-25T00:00:00Z" {
		t.Fatalf("expires_at = %q", got.ExpiresAt)
	}
}
