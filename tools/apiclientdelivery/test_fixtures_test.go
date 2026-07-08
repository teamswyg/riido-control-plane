package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/requirements"
)

func apiClientDeliveryFixture() manifest {
	return manifest{
		SchemaVersion: requirements.ManifestSchema, ID: "api-client-delivery",
		Title: "API Client Delivery", RiidoTask: "RIID-test",
		GeneratedDoc: "docs/api-client.md", Workflow: ".github/workflows/api.yml",
		Evidence: "api-client-evidence", RiskEvidence: "risk.json",
		Sources:   []sourceRef{{Name: "contract", Path: "contracts/api.json"}},
		Owners:    []owner{{Name: "control-plane", Owns: "contract", DoesNotOwn: "client"}},
		Delivery:  delivery{Workflow: "delivery.yml", PackageMode: "generated", DeliverMode: "artifact"},
		Branch:    branchRule{Source: "main", Rule: "generated only", Example: "main", SecretGate: "no secrets"},
		Lifecycle: []string{"generated"}, Generator: generator{ReactQuery: "on", Handoff: "on"},
		Figma:        []figmaContext{{ID: "runtime", NodeIDs: []string{"1:2"}, GeneratedPaths: []string{"docs/api.md"}, Rule: "read only"}},
		ModelCatalog: modelCatalog{Policy: "provider", Rendering: "label", FixtureRule: "no hardcode"},
		Required:     []string{"API Client Delivery"},
		Forbidden:    []string{"forbidden stale phrase"},
		Checks:       []sourceCheck{{Name: "server", File: "internal/server.go", Contains: []string{"handler"}}},
		Loop:         loopRecord{"observe", "hypothesis", "execute", "evaluate", "retro"},
		NonGoals:     []string{"edit client"},
	}
}

func writeAPIClientDeliveryRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeAPIClientFile(t, filepath.Join(repo, "go.mod"), "module example.com/api\n")
	writeAPIClientFile(t, filepath.Join(repo, "contracts/api.json"), "{}")
	writeAPIClientFile(t, filepath.Join(repo, "internal/server.go"), "package internal\n// handler\n")
	writeAPIClientJSON(t, filepath.Join(repo, "risk.json"), riskManifest{
		Local: []riskEvidence{{Risk: "drift", Test: "go test", Proves: "stable"}},
	})
	writeAPIClientJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeAPIClientJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeAPIClientFile(t, path, string(body))
}

func writeAPIClientFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
