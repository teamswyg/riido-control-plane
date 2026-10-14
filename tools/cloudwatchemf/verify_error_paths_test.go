package main

import (
	"strings"
	"testing"
)

func TestVerifyRejectsMissingDimension(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.RequiredDimensions = []string{"missing_dimension"}
	shape, err := buildEMFShape()
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyDimensions(m, shape); err == nil || !strings.Contains(err.Error(), "missing_dimension") {
		t.Fatalf("verifyDimensions error = %v, want missing_dimension", err)
	}
}

func TestVerifyRejectsMissingJSONFieldScopeAndMetricUnit(t *testing.T) {
	t.Parallel()
	shape, err := buildEMFShape()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		mut  func(manifest) manifest
		run  func(manifest, emfShape) error
		want string
	}{
		{"field", func(m manifest) manifest { m.RequiredJSONFields = []string{"missing"}; return m }, verifyJSONFields, "missing"},
		{"scope", func(m manifest) manifest {
			m.RequiredScopes = []requiredScope{{Field: "metric_scope_schema_version", Value: "nope"}}
			return m
		}, verifyScopes, "metric_scope_schema_version"},
		{"unit", func(m manifest) manifest {
			m.RequiredMetricUnit = []requiredUnit{{Name: "tasks_total", Unit: "Milliseconds"}}
			return m
		}, verifyMetricUnits, "tasks_total"},
	}
	for _, tc := range cases {
		if err := tc.run(tc.mut(testManifest()), shape); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %s", tc.name, err, tc.want)
		}
	}
}

func TestDecodeEMFShapeRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	if _, err := decodeEMFShape([]byte(`{`)); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestVerifyDocReportsReadError(t *testing.T) {
	t.Parallel()
	if err := verifyDoc(t.TempDir(), testManifest()); err == nil ||
		!strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("verifyDoc error = %v, want read generated doc", err)
	}
}
