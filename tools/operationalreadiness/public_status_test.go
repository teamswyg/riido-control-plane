package main

import (
	"strings"
	"testing"
	"time"
)

func TestOperationalReadinessPublicStatusIsRedacted(t *testing.T) {
	m := loadManifestForTest(t)
	got := newEvidenceAt(m, time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC))
	status := got.PublicStatus
	if status.Visibility != "public_aggregate" || status.EndpointDetails != "redacted" {
		t.Fatalf("public status must be aggregate/redacted: %+v", status)
	}
	if status.RawLogsIncluded || status.SecretsIncluded {
		t.Fatalf("public status leaked raw data flags: %+v", status)
	}
	if status.P0CycleCount == 0 || status.P0PartialCount == 0 || status.Overall != "degraded" {
		t.Fatalf("public status should expose open P0 QA risk: %+v", status)
	}
	if status.GeneratedAt != got.GeneratedAt || status.ExpiresAt != got.ExpiresAt {
		t.Fatalf("public status freshness mismatch: %+v evidence=%+v", status, got)
	}
	if status.EvidenceTTLHours != got.EvidenceTTLHours {
		t.Fatalf("public status ttl = %d", status.EvidenceTTLHours)
	}
}

func TestOperationalReadinessPublicStatusDocHasNoSecretMarkers(t *testing.T) {
	m := loadManifestForTest(t)
	e := newEvidenceAt(m, time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC))
	doc := renderDoc(m, e)
	for _, marker := range []string{"eyJ", "AKIA", "device_secret", "RIIDO_DEVICE_SECRET"} {
		if strings.Contains(doc, marker) {
			t.Fatalf("public status doc contains forbidden marker %q", marker)
		}
	}
	if !strings.Contains(doc, "## Public QA Status") {
		t.Fatal("public status section missing")
	}
	if strings.Contains(doc, "generated at: `2026-06-30T10:00:00Z`") {
		t.Fatalf("generated architecture doc should not contain runtime freshness: %s", doc)
	}
}
