package main

import (
	"path/filepath"
	"testing"
)

func testOpenAPIPath() string {
	return filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json")
}

func testContractBasePath() string {
	return filepath.Join("..", "..", "contracts", "ai-agent-client")
}

func loadTestOpenAPI(t *testing.T) openAPISpec {
	t.Helper()
	spec, err := loadOpenAPI(testOpenAPIPath())
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	return spec
}
