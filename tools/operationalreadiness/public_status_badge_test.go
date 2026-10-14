package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusBadge(t *testing.T) {
	out := t.TempDir() + "/status-badge.json"
	err := run(options{
		Repo:                  "../..",
		Manifest:              defaultManifest,
		CheckDoc:              true,
		PublicStatusBadgeJSON: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got publicStatusBadge
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 || got.Label != "riido qa" {
		t.Fatalf("badge identity = %+v", got)
	}
	if !strings.Contains(got.Message, "degraded") || got.Color == "" {
		t.Fatalf("badge status = %+v", got)
	}
	if strings.Contains(string(data), "eyJ") || strings.Contains(string(data), "RIIDO_DEVICE_SECRET") {
		t.Fatalf("badge leaked secret marker: %s", data)
	}
}
