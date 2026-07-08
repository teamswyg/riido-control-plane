package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func webFrontendAPIFixture() manifest {
	return manifest{
		ID: "control-plane-web-frontend-api", Title: "Web Frontend API",
		GeneratedDoc: "docs/web-frontend-api.md", Workflow: ".github/workflows/web.yml",
		EvidenceArtifact:  "web-frontend-api-evidence",
		OwnerPackages:     []string{"internal/riidoaiserver"},
		RuntimeConfigKeys: []string{"RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS"},
		CORSCases: []corsCase{{
			Name: "allowed-actual", AllowedOrigins: []string{"http://localhost:3000"},
			Method: "GET", Path: "/healthz", Origin: "http://localhost:3000",
			WantStatus: 200, WantAllowOrigin: "http://localhost:3000",
		}},
		SourceChecks: []sourceCheck{{Name: "cors", File: "internal/cors.go", Contains: []string{"WebAllowedOrigins"}}},
		Loop:         evidenceLoop{"observe", "hypothesis", "execute", "evaluate", "retro"},
	}
}

func writeWebFrontendRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeWebFrontendFile(t, filepath.Join(repo, "go.mod"), "module example.com/web\n")
	writeWebFrontendFile(t, filepath.Join(repo, "internal/cors.go"), "package internal\n// WebAllowedOrigins\n")
	writeWebFrontendJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeWebFrontendJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeWebFrontendFile(t, path, string(body))
}

func writeWebFrontendFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
