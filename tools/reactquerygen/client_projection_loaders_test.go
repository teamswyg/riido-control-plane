package main

import (
	"encoding/json"
	"os"
	"testing"
)

func loadContractClientProjection(t *testing.T, path string) clientProjection {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc struct {
		ClientModules []clientModuleMetadata  `json:"client_modules"`
		Operations    []testContractOperation `json:"operations"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return clientProjection{
		modules:    canonicalJSON(t, doc.ClientModules),
		operations: clientMetadataByOperation(t, path, doc.Operations),
	}
}

func loadOpenAPIClientProjection(t *testing.T, path string) clientProjection {
	t.Helper()
	spec, err := loadOpenAPI(path)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	var operations []testContractOperation
	for _, byMethod := range spec.Paths {
		for _, operation := range byMethod {
			operations = append(operations, testContractOperation{
				OperationID: operation.OperationID,
				Client:      operation.Client,
			})
		}
	}
	return clientProjection{
		modules:    canonicalJSON(t, spec.ClientModules),
		operations: clientMetadataByOperation(t, path, operations),
	}
}
