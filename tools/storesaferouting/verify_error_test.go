package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyCaseNamesRejectsResultDrift(t *testing.T) {
	t.Parallel()
	cases := []routingCase{{Name: "expected"}}
	assertStoreSafeRoutingError(t, verifyCaseNames(cases, nil), "routing result count=0")
	assertStoreSafeRoutingError(t,
		verifyCaseNames(cases, []routingEvidence{{}}),
		"empty case name")
	assertStoreSafeRoutingError(t,
		verifyCaseNames(cases, []routingEvidence{{Name: "other"}}),
		"missing routing evidence case expected")
}

func TestVerifyDocRejectsMissingAndStaleDoc(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	m := manifest{GeneratedDoc: "docs/generated.md", Title: "Expected"}
	assertStoreSafeRoutingError(t, verifyDoc(root, m), "read generated doc")
	docPath := filepath.Join(root, filepath.FromSlash(m.GeneratedDoc))
	if err := writeText(docPath, "stale"); err != nil {
		t.Fatal(err)
	}
	assertStoreSafeRoutingError(t, verifyDoc(root, m), "is stale")
}

func TestVerifySourceChecksRejectsMissingSourceAndNeedle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	check := sourceCheck{Name: "source", File: "source.go", Contains: []string{"needle"}}
	assertStoreSafeRoutingError(t, verifySourceChecks(root, []sourceCheck{check}), "read source check source")
	if err := writeText(filepath.Join(root, "source.go"), "haystack"); err != nil {
		t.Fatal(err)
	}
	assertStoreSafeRoutingError(t, verifySourceChecks(root, []sourceCheck{check}), "missing \"needle\"")
}

func assertStoreSafeRoutingError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want contains %q", err, want)
	}
}
