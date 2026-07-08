package main

import (
	"strings"
	"testing"
)

func TestConfigReferenceVerifyEntryErrors(t *testing.T) {
	t.Parallel()
	cases := []entry{
		{},
		{Name: "RIIDO_TOKEN", Default: "unset", Owner: "owner", Sensitivity: "public", Meaning: "meaning"},
		{Name: "RIIDO_ENV", Default: "unset", Owner: "owner", Meaning: "meaning"},
	}
	for _, entry := range cases {
		if err := verifyEntry(entry); err == nil {
			t.Fatalf("expected verifyEntry error for %+v", entry)
		}
	}
}

func TestConfigReferenceVerifyManifestShapeErrors(t *testing.T) {
	t.Parallel()
	if err := verifyManifestShape(manifest{}); err == nil {
		t.Fatal("expected required manifest fields error")
	}
	m := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	m.Loop.Execute = ""
	if err := verifyManifestShape(m); err == nil || !strings.Contains(err.Error(), "loop") {
		t.Fatalf("expected loop error, got %v", err)
	}
}

func TestConfigReferenceVerifyEnvParityErrors(t *testing.T) {
	t.Parallel()
	duplicate := testManifest("cmd/app", testEntry("RIIDO_ENV"), testEntry("RIIDO_ENV"))
	if err := verifyEnvParity(duplicate, map[string]bool{"RIIDO_ENV": true}); err == nil {
		t.Fatal("expected duplicate env error")
	}
	missingRead := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	if err := verifyEnvParity(missingRead, map[string]bool{}); err == nil {
		t.Fatal("expected manifest env not read error")
	}
	missingManifest := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	sourceNames := map[string]bool{"RIIDO_ENV": true, "AWS_REGION": true}
	if err := verifyEnvParity(missingManifest, sourceNames); err == nil {
		t.Fatal("expected source env missing from manifest error")
	}
}
