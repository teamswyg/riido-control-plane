package main

import "testing"

func TestVerifyMatrixRejectsEntryDrift(t *testing.T) {
	m := smokeFixtureManifest()
	ops := map[string]generatedOperation{
		"v1.foo": {"GET", "/v1/foo"},
		"v2.bar": {"POST", "/v2/bar"},
	}
	for name, matrix := range map[string]smokeMatrix{
		"schema": {SchemaVersion: "other"},
		"unknown": {SchemaVersion: m.SmokeSchemaVersion, Entries: []smokeEntry{
			{"v9.nope", "GET", "/v9/nope", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
		}},
		"drift": {SchemaVersion: m.SmokeSchemaVersion, Entries: []smokeEntry{
			{"v1.foo", "POST", "/v1/foo", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
		}},
		"unsorted": {SchemaVersion: m.SmokeSchemaVersion, Entries: []smokeEntry{
			{"v2.bar", "POST", "/v2/bar", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV2"}},
			{"v1.foo", "GET", "/v1/foo", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
		}},
		"missing": {SchemaVersion: m.SmokeSchemaVersion, Entries: []smokeEntry{
			{"v1.foo", "GET", "/v1/foo", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := verifyMatrix(m, ops, matrix); err == nil {
				t.Fatalf("expected matrix error")
			}
		})
	}
}

func TestVerifyEvidenceTestsRejectsWrongCoverage(t *testing.T) {
	m := smokeFixtureManifest()
	for name, entry := range map[string]smokeEntry{
		"unknown": {"v1.foo", "GET", "/v1/foo", []string{"missing-test"}},
		"empty":   {"v1.foo", "GET", "/v1/foo", nil},
		"v1":      {"v1.foo", "GET", "/v1/foo", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV2"}},
		"v2":      {"v2.bar", "POST", "/v2/bar", []string{"TestHTTPAIAgentClientGeneratedEndpointSmokeV1"}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := verifyEvidenceTests(m, entry); err == nil {
				t.Fatalf("expected evidence-test error")
			}
		})
	}
	if hasString([]string{"a", "b"}, "z") {
		t.Fatalf("hasString reported a missing value")
	}
}
