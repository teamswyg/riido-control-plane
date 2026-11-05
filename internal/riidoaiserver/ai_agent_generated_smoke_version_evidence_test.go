package riidoaiserver

import (
	"strings"
	"testing"
)

func verifyGeneratedSmokeVersionEvidence(t *testing.T, entry aiAgentGeneratedEndpointSmokeEntry) {
	t.Helper()
	if strings.HasPrefix(entry.GeneratedPath, "v2.") && !containsString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV2") && !containsString(entry.EvidenceTests, "TestMessageReactionHTTPAuthorizesAndPersistsActor") {
		t.Fatalf("v2 matrix entry must cite v2 smoke evidence: %+v", entry)
	}
	if strings.HasPrefix(entry.GeneratedPath, "v3.") && !containsString(entry.EvidenceTests, "TestHTTPAIAgentClientV3ThreadHistory") {
		t.Fatalf("v3 matrix entry must cite canonical history evidence: %+v", entry)
	}
	if !strings.HasPrefix(entry.GeneratedPath, "v2.") && !strings.HasPrefix(entry.GeneratedPath, "v3.") && !containsString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV1") {
		t.Fatalf("v1 matrix entry must cite v1 smoke evidence: %+v", entry)
	}
}
