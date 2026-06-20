package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "readme-evidence.json")
	if err := runWithOptions("../..", defaultManifest, false, true, out); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "verified" || got.DocLinkCount == 0 || got.EndpointCount == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}
