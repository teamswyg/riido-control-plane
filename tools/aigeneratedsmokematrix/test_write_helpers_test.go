package main

import (
	"encoding/json"
	"os"
	"testing"
)

func smokeFixtureMatrix(m manifest) smokeMatrix {
	return smokeMatrix{SchemaVersion: m.SmokeSchemaVersion, Entries: []smokeEntry{
		{"v1.foo", "GET", "/v1/foo", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
		{"v2.bar", "POST", "/v2/bar", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV2"}},
	}}
}

func mustJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}
