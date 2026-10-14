package main

import (
	"strings"
	"testing"
)

func TestVerifyFieldsRejectsMissingJSONField(t *testing.T) {
	t.Parallel()
	m := testManifest()
	result := adapterResult{Fields: map[string]bool{"schema_version": true}}
	if err := verifyFields(m, result); err == nil || !strings.Contains(err.Error(), "generated_at") {
		t.Fatalf("verifyFields error = %v, want generated_at", err)
	}
}

func TestVerifyRejectsMissingBreakdownRows(t *testing.T) {
	t.Parallel()
	m := testManifest()
	root := writeTestRepo(t, m, "needle")
	result, err := exerciseAdapter(m)
	if err != nil {
		t.Fatal(err)
	}
	result.HTTPBreakdownRows = 0
	if err := verify(root, m, result, false); err == nil || !strings.Contains(err.Error(), "breakdown") {
		t.Fatalf("verify error = %v, want breakdown", err)
	}
}

func TestVerifyRejectsStatusDrift(t *testing.T) {
	t.Parallel()
	m := testManifest()
	result, err := exerciseAdapter(m)
	if err != nil {
		t.Fatal(err)
	}
	result.MissingScopeStatus = 200
	if err := verifyStatuses(m, result); err == nil || !strings.Contains(err.Error(), "missing_scope") {
		t.Fatalf("verifyStatuses error = %v, want missing_scope", err)
	}
}

func TestDecodeSnapshotRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	if _, _, err := decodeSnapshot([]byte(`{`)); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestExerciseAdapterRejectsUnexpectedMetricsBody(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.Endpoint.Path = "/missing"
	if _, err := exerciseAdapter(m); err == nil {
		t.Fatal("expected adapter decode error")
	}
}

func TestVerifyDocReportsReadError(t *testing.T) {
	t.Parallel()
	m := testManifest()
	if err := verifyDoc(t.TempDir(), m); err == nil || !strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("verifyDoc error = %v, want read generated doc", err)
	}
}
