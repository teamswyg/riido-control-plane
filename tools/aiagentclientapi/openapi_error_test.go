package main

import (
	"path/filepath"
	"testing"
)

func TestLoadOpenAPIOperationsRejectsMissingAndInvalid(t *testing.T) {
	t.Parallel()
	if _, err := loadOpenAPIOperations(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing openapi error")
	}
	path := filepath.Join(t.TempDir(), "openapi.json")
	if err := writeText(path, "{"); err != nil {
		t.Fatal(err)
	}
	if _, err := loadOpenAPIOperations(path); err == nil {
		t.Fatal("expected invalid openapi error")
	}
}

func TestLoadOpenAPIOperationsCountsOnlyHTTPMethods(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "openapi.json")
	body := `{"paths":{
		"/v1/client/ai-agent/bootstrap":{"get":{"x-riido-client":{"generated_path":"v1.bootstrap"}},"parameters":{}},
		"/v2/client/workspaces/{workspace_id}/ai-agent/bootstrap":{"post":{"x-riido-client":{"generated_path":"v2.bootstrap"}}},
		"/outside":{"options":{"x-riido-client":{"generated_path":"ignored"}}}
	}}`
	if err := writeText(path, body); err != nil {
		t.Fatal(err)
	}
	ops, err := loadOpenAPIOperations(path)
	if err != nil {
		t.Fatal(err)
	}
	if ops.Counts.Total != 2 || ops.Counts.V1 != 1 || ops.Counts.V2 != 1 {
		t.Fatalf("counts = %+v", ops.Counts)
	}
	if _, ok := ops.Generated["ignored"]; ok {
		t.Fatal("non-HTTP method should not be captured")
	}
}
