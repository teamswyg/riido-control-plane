package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryReadmeBehaviorGolden(t *testing.T) {
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
	if got.SchemaVersion != evidenceSchema || got.ID != "control-plane-repository-readme" {
		t.Fatalf("unexpected evidence identity: %+v", got)
	}
	if got.Status != "verified" || got.Manifest != defaultManifest || got.GeneratedDoc != generatedDoc {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.FragmentCount != 7 || got.DocLinkCount != 14 || got.EndpointCount != 15 {
		t.Fatalf("unexpected evidence counts: %+v", got)
	}
	if got.VerificationCount != 9 || got.RuntimeCDNoteCount != 4 || got.RequiredMarkerCount != 9 {
		t.Fatalf("unexpected verification evidence: %+v", got)
	}
	if !completeLoop(got.Loop) {
		t.Fatalf("evidence loop is incomplete: %+v", got.Loop)
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
