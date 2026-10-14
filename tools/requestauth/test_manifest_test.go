package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testManifest() manifest {
	surfaces := make([]surface, 0, len(requiredSurfaces))
	for _, name := range requiredSurfaces {
		surfaces = append(surfaces, surface{Name: name, Role: "role"})
	}
	transports := make([]transport, 0, len(requiredTransports))
	for _, id := range requiredTransports {
		transports = append(transports, transport{ID: id, Value: id})
	}
	groups := make([]ruleGroup, 0, len(requiredRuleGroups))
	for id, rules := range requiredRuleGroups {
		groups = append(groups, ruleGroup{ID: id, Rules: append([]string{}, rules...)})
	}
	return manifest{
		SchemaVersion: manifestSchema, ID: expectedID, Title: "Request Authorization",
		RiidoTask: expectedTask, GeneratedDoc: "out/request.md", Workflow: "workflow.yml",
		EvidenceArtifact: "evidence", OwnerPackage: "internal/riidoaiserver",
		Surfaces: surfaces, Resources: append([]string{}, requiredResources...),
		TokenTransports: transports, RuntimeConfigKeys: append([]string{}, requiredRuntimeConfigKeys...),
		ExternalContractVersions: append([]string{}, requiredContractVersions...),
		RuleGroups:               groups, SourceChecks: []sourceCheck{{Name: "source", File: "source.txt", Contains: []string{"needle"}}},
		Loop: loop{Observation: "o", Hypothesis: "h", Execute: "x", Evaluate: "e", Retrospective: "r"},
	}
}

func writeTestRepo(t *testing.T, m manifest, source string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module test\n")
	writeFile(t, root, "source.txt", source)
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "manifest.json", string(body))
	return root
}

func writeFile(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
