package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func clientMetadataByOperation(t *testing.T, path string, operations []testContractOperation) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, operation := range operations {
		operationID := strings.TrimSpace(operation.OperationID)
		if operationID == "" {
			t.Fatalf("%s has operation without id", path)
		}
		if strings.TrimSpace(operation.Client.Module) == "" {
			t.Fatalf("%s operation %s missing client.module", path, operationID)
		}
		if len(operation.Client.FacadePath) == 0 {
			t.Fatalf("%s operation %s missing client.facade_path", path, operationID)
		}
		operation.Client.GeneratedPath = generatedPathFromClient(operation.Client)
		out[operationID] = canonicalJSON(t, operation.Client)
	}
	return out
}

func canonicalJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return string(data)
}
