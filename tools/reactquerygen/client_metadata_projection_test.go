package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAIAgentClientMetadataFlowsThroughContractProjections(t *testing.T) {
	base := testContractBasePath()
	dsl := loadContractClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.dsl.riido.json"))
	ir := loadContractClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.ir.riido.json"))
	openapi := loadOpenAPIClientProjection(t, filepath.Join(base, "control-plane-ai-agent-client.openapi.json"))
	assertClientProjectionMatches(t, "IR", ir, dsl)
	assertClientProjectionMatches(t, "OpenAPI", openapi, dsl)
	assertRequiredClientMetadataFields(t, dsl)
}

func assertClientProjectionMatches(t *testing.T, label string, got, want clientProjection) {
	t.Helper()
	if got.modules != want.modules {
		t.Fatalf("%s client modules = %s, want %s", label, got.modules, want.modules)
	}
	for operationID, wantMetadata := range want.operations {
		if gotMetadata, ok := got.operations[operationID]; !ok || gotMetadata != wantMetadata {
			t.Fatalf("%s client metadata for %s = %q, want %q", label, operationID, gotMetadata, wantMetadata)
		}
	}
	if len(got.operations) != len(want.operations) {
		t.Fatalf("%s metadata count = %d, want %d", label, len(got.operations), len(want.operations))
	}
}

func assertRequiredClientMetadataFields(t *testing.T, dsl clientProjection) {
	t.Helper()
	for operationID, required := range requiredClientMetadataFields {
		metadata := dsl.operations[operationID]
		for _, field := range required {
			if !strings.Contains(metadata, field) {
				t.Fatalf("%s metadata missing %s in %s", operationID, field, metadata)
			}
		}
	}
}
