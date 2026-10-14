package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-cloudwatch-emf.v1",
		ID:               "control-plane-cloudwatch-emf",
		Title:            "Control Plane CloudWatch EMF",
		GeneratedDoc:     "out/cloudwatch-emf.md",
		Workflow:         ".github/workflows/cloudwatch-emf.yml",
		EvidenceArtifact: "cloudwatch-emf-evidence",
		OwnerPackage:     "internal/riidoaiserver",
		SourceChecks: []sourceCheck{
			{Name: "source", File: "source.txt", Contains: []string{"needle"}},
		},
		RequiredDimensions: []string{"service"},
		RequiredJSONFields: []string{"schema_version", "service", "http_transactions", "store_operations"},
		RequiredScopes: []requiredScope{
			{Field: "metric_scope_schema_version", Value: "riido-metric-scope.v1"},
		},
		RequiredMetricUnit: []requiredUnit{{Name: "tasks_total", Unit: "Count"}},
		Loop: evidenceLoop{
			Observation: "o", Hypothesis: "h", Execute: "x", Evaluate: "e", Retrospective: "r",
		},
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
