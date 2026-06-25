package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestControlPlaneHighTrafficAuditVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.SurfaceCount < 7 || got.CandidateCount != got.SurfaceCount {
		t.Fatalf("audit coverage = %+v", got)
	}
}

func TestControlPlaneHighTrafficAuditRejectsUnsafePprof(t *testing.T) {
	m := loadManifestForTest(t)
	m.PprofCommands = []string{"go tool pprof http://0.0.0.0:6060/debug/pprof/profile"}
	if err := verifyCommands(m); err == nil {
		t.Fatal("expected unsafe pprof command to fail")
	}
}

func TestControlPlaneHighTrafficAuditRejectsMissingPattern(t *testing.T) {
	m := loadManifestForTest(t)
	m.Surfaces[0].Patterns = []string{"definitely_missing_hot_path_marker"}
	if err := verifySurfaces("../..", m.Surfaces); err == nil {
		t.Fatal("expected missing pattern to fail")
	}
}

func loadManifestForTest(t *testing.T) manifest {
	t.Helper()
	var m manifest
	if err := readJSON("../../"+defaultManifest, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func readEvidence(t *testing.T, path string) evidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
