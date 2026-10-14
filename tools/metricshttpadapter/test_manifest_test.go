package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/requirements"
)

func testManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-metrics-http-adapter.v1",
		ID:               requirements.ExpectedID,
		Title:            "Metrics HTTP Adapter",
		GeneratedDoc:     "out/metrics.md",
		Workflow:         requirements.Workflow,
		EvidenceArtifact: requirements.EvidenceArtifact,
		OwnerPackage:     "internal/riidoaiserver",
		Endpoint: endpointContract{
			Method: "GET", Path: "/metrics", Resource: "metrics", Action: "read",
		},
		SourceChecks: []sourceCheck{{Name: "source", File: "source.txt", Contains: []string{"needle"}}},
		RequiredFields: []string{
			"schema_version", "generated_at", "tasks_total", "assignments_total",
			"assignments_by_state", "poll_requests_total", "poll_actions_total",
			"agent_events_total", "task_events_total", "event_append_latency_samples_total",
			"http_transactions", "store_operations",
		},
		RequiredStatuses: []statusContract{
			{Case: "authorized", Status: 200},
			{Case: "missing_scope", Status: 403},
			{Case: "store_unconfigured", Status: 503},
		},
		Loop: evidenceLoop{Observation: "o", Hypothesis: "h", Execute: "x", Evaluate: "e", Retrospective: "r"},
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
