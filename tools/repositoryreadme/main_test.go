package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	if got.Status != "verified" || got.FragmentCount == 0 || got.DocLinkCount == 0 || got.EndpointCount == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.RequiredMarkerCount < 9 {
		t.Fatalf("public entrypoint markers missing: %+v", got)
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "actions/workflows/operational-readiness.yml/badge.svg") {
		t.Fatalf("README public QA status badge missing")
	}
	if !strings.Contains(string(readme), "https://teamswyg.github.io/riido-control-plane/") {
		t.Fatalf("README public QA Pages URL missing")
	}
}
