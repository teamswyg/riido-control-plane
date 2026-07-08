package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/providerstatus/requirements"
)

func writeProviderStatusTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeProviderStatusManifest(t *testing.T, path string, m manifest) {
	t.Helper()
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeProviderStatusTestFile(t, path, string(body))
}

func minimalProviderStatusManifest() manifest {
	return manifest{
		SchemaVersion:    requirements.ManifestSchema,
		ID:               requirements.ExpectedID,
		Title:            "Provider Status",
		RiidoTask:        requirements.ExpectedTask,
		GeneratedDoc:     "docs/generated.md",
		Workflow:         ".github/workflows/provider-status.yml",
		EvidenceArtifact: "provider-status-evidence",
		OwnerPackage:     "internal/riidoaiserver",
		Authorization:    []authRule{{Action: "read", Scope: "scope"}},
		Loop: evidenceLoop{
			Observation:   "observation",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retrospective",
		},
	}
}

func completeProviderStatusManifest() manifest {
	m := minimalProviderStatusManifest()
	for _, name := range requirements.Surfaces {
		m.Surfaces = append(m.Surfaces, surface{Name: name, Role: "role"})
	}
	for _, value := range requirements.RoutingStatuses {
		m.RoutingStatuses = append(m.RoutingStatuses, valueRecord(value))
	}
	for _, value := range requirements.DistributionChannels {
		m.DistributionChannels = append(m.DistributionChannels, valueRecord(value))
	}
	for _, id := range requirements.ValidationRules {
		m.ValidationRules = append(m.ValidationRules, rule{ID: id, Rule: "rule"})
	}
	for _, id := range requirements.RoutingRules {
		m.RoutingRules = append(m.RoutingRules, rule{ID: id, Rule: "rule"})
	}
	return m
}

func valueRecord(name string) value {
	return value{Value: name, Owner: "owner"}
}
