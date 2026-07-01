package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusJSON(t *testing.T) {
	out := t.TempDir() + "/public-status.json"
	err := run(options{
		Repo:             "../..",
		Manifest:         defaultManifest,
		CheckDoc:         true,
		PublicStatusJSON: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got publicStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Visibility != "public_aggregate" || got.EndpointDetails != "redacted" {
		t.Fatalf("public status json not redacted: %+v", got)
	}
	if got.GeneratedAt == "" || got.ExpiresAt == "" || got.EvidenceTTLHours == 0 {
		t.Fatalf("public status json freshness missing: %+v", got)
	}
	if strings.Contains(string(data), "eyJ") || strings.Contains(string(data), "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status json leaked secret marker: %s", data)
	}
}
