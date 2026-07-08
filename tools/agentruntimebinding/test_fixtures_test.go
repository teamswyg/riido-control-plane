package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/agentruntimebinding/requirements"
)

func agentRuntimeBindingFixture() manifest {
	return manifest{
		SchemaVersion:    requirements.ManifestSchema,
		ID:               requirements.ExpectedID,
		Title:            "Agent Runtime Binding",
		RiidoTask:        requirements.ExpectedTask,
		GeneratedDoc:     "docs/agent-runtime-binding.md",
		Workflow:         ".github/workflows/agent-runtime-binding.yml",
		EvidenceArtifact: "agent-runtime-binding-evidence",
		OwnerPackage:     "internal/riidoaiserver",
		BindingFields: []field{
			{Name: "agent_id", Required: true},
			{Name: "daemon_id", Required: true},
			{Name: "device_id", Required: false},
			{Name: "runtime_id", Required: true},
			{Name: "runtime_provider", Required: true},
		},
		BindingRules: bindingRulesFixture(requirements.RequiredRules),
		DeviceRules:  bindingRulesFixture(requirements.RequiredDeviceRules),
		SourceChecks: []sourceCheck{{Name: "registry", File: "internal/registry.go", Contains: []string{"validate"}}},
		Loop:         evidenceLoop{"observe", "hypothesis", "execute", "evaluate", "retro"},
		NonGoals:     []string{"queue"},
	}
}

func bindingRulesFixture(ids []string) []rule {
	rules := make([]rule, 0, len(ids))
	for _, id := range ids {
		rules = append(rules, rule{ID: id, Kind: "policy", Rule: "must hold"})
	}
	return rules
}

func writeAgentRuntimeBindingRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeAgentRuntimeBindingFile(t, filepath.Join(repo, "go.mod"), "module example.com/runtime\n")
	writeAgentRuntimeBindingFile(t, filepath.Join(repo, "internal/registry.go"), "package internal\n// validate\n")
	writeAgentRuntimeBindingJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeAgentRuntimeBindingJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeAgentRuntimeBindingFile(t, path, string(body))
}

func writeAgentRuntimeBindingFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
