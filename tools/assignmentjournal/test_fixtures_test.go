package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/assignmentjournal/requirements"
)

func assignmentJournalFixture() manifest {
	m := manifest{
		SchemaVersion: requirements.ManifestSchema, ID: requirements.ExpectedID,
		Title: "Assignment Journal", RiidoTask: requirements.ExpectedTask,
		GeneratedDoc: "docs/assignment.md", Workflow: ".github/workflows/assignment.yml",
		EvidenceArtifact: "assignment-evidence", OwnerPackage: "internal/riidoaiserver",
		Loop:         evidenceLoop{"observe", "hypothesis", "execute", "evaluate", "retro"},
		NonGoals:     []string{"no runtime mutation"},
		SourceChecks: []sourceCheck{{Name: "store", File: "internal/store.go", Contains: []string{"claim", "lease"}}},
	}
	for _, name := range requirements.RequiredPorts {
		m.Ports = append(m.Ports, surface{Name: name, Role: "port"})
	}
	for _, name := range requirements.RequiredRecords {
		m.Records = append(m.Records, surface{Name: name, Role: "record"})
	}
	for _, id := range requirements.RequiredReplayRules {
		m.ReplayRules = append(m.ReplayRules, rule{ID: id, Kind: "rule", Rule: "required"})
	}
	for _, name := range requirements.RequiredConstants {
		m.VersionConstants = append(m.VersionConstants, constant{Name: name, Value: "v1"})
	}
	return m
}

func writeAssignmentJournalRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeAssignmentFile(t, filepath.Join(repo, "go.mod"), "module example.com/assignment\n")
	writeAssignmentFile(t, filepath.Join(repo, "internal/store.go"), "package internal\n// claim lease\n")
	writeAssignmentJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeAssignmentJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeAssignmentFile(t, path, string(body))
}

func writeAssignmentFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
