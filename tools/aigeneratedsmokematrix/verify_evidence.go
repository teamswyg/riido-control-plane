package main

import "fmt"

func verifyEvidenceTests(m manifest, entry smokeEntry) error {
	allowed := stringSet(m.RequiredEvidenceTests)
	for _, test := range entry.EvidenceTests {
		if !allowed[test] {
			return fmt.Errorf("unknown evidence test %q for %s", test, entry.GeneratedPath)
		}
	}
	if len(entry.EvidenceTests) == 0 {
		return fmt.Errorf("missing evidence test for %s", entry.GeneratedPath)
	}
	if wantsV2(entry) && !hasString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV2") {
		return fmt.Errorf("v2 entry missing v2 evidence: %s", entry.GeneratedPath)
	}
	if !wantsV2(entry) && !hasString(entry.EvidenceTests, "TestHTTPAIAgentClientGeneratedEndpointSmokeV1") {
		return fmt.Errorf("v1 entry missing v1 evidence: %s", entry.GeneratedPath)
	}
	return nil
}

func hasString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
