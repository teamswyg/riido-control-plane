package main

import (
	"strings"
	"testing"
)

func verifyRuntimeEndpointLabelProjection(t *testing.T, sourceCoverage figmaSourceCoverageManifest, docText string) {
	t.Helper()
	if !figmaRuntimeEndpointEvidenceFound(sourceCoverage.VerifiedEvidenceNodes) {
		t.Fatal("source coverage must register runtime settings endpoint-looking label node-id=129:17930")
	}
	runtimeEntry := figmaRuntimeSettingsSourceEntry(t, sourceCoverage.Entries)
	facts := strings.Join(runtimeEntry.CoveredFacts, "\n")
	requireDocMentions(t, facts, "runtime settings facts", []string{
		"node-id=129:17930",
		"not a canonical base URL",
		"generated path",
		"live host export",
	})
	requireDocMentions(t, docText, "endpoint-looking label boundary", []string{
		"node-id=129:17930",
		"not a canonical base URL",
		"generated path",
		"live host export",
	})
}

func figmaRuntimeEndpointEvidenceFound(nodes []figmaSourceCoverageNode) bool {
	for _, node := range nodes {
		if node.NodeID == "129:17930" && strings.Contains(strings.ToLower(node.Name), "endpoint") {
			return true
		}
	}
	return false
}

func figmaRuntimeSettingsSourceEntry(t *testing.T, entries []figmaSourceCoverageEntry) figmaSourceCoverageEntry {
	t.Helper()
	for _, entry := range entries {
		if entry.NodeID == "162:23090" {
			return entry
		}
	}
	t.Fatal("source coverage runtime settings entry 162:23090 is missing")
	return figmaSourceCoverageEntry{}
}
