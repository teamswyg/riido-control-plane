package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func migrationLedgerFixture() manifest {
	sections := []section{
		{2, "Goal", "overview", []string{"RIID-4638"}},
		{2, "Source In The Private Repository", "overview", []string{"RIID-4638"}},
		{2, "Target Boundary", "overview", []string{"RIID-4638"}},
		{2, "Migration Order", "overview", []string{"RIID-4638"}},
		{2, "Current Migration Slices", "overview", []string{"RIID-4638"}},
		{2, "Validation Gates", "validation", []string{"- gate", "- gate", "- gate", "- gate", "- gate"}},
		{2, "Infra Boundary", "overview", []string{"RIID-4638"}},
		{2, "Contract Boundary", "overview", []string{"RIID-4638"}},
	}
	for i := 0; i < 80; i++ {
		sections = append(sections, section{3, "RIID-slice", "slice", []string{"RIID-4638"}})
	}
	return manifest{
		SchemaVersion: manifestSchema, ID: expectedID, Title: "Migration Ledger",
		RiidoTask: expectedTask, GeneratedDoc: "docs/migration.md",
		Workflow: ".github/workflows/migration.yml", EvidenceArtifact: "migration-evidence",
		Intro: []string{"intro"}, Sections: sections, Assertions: []string{"assert"},
		Loop: evidenceLoop{
			Observation: "observe", Hypothesis: "hypothesis", Execute: "execute",
			Evaluate: "evaluate", Retrospective: "retro",
		},
	}
}

func writeMigrationLedgerRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeMigrationFile(t, filepath.Join(repo, "go.mod"), "module example.com/migration\n")
	writeMigrationJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeMigrationJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeMigrationFile(t, path, string(body))
}

func writeMigrationFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
