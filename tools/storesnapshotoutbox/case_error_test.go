package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCaseRejectsUnknownKind(t *testing.T) {
	t.Parallel()
	if _, err := runCase(caseSpec{Name: "bad", Kind: "unknown"}); err == nil {
		t.Fatal("runCase accepted unknown kind")
	}
	if _, err := verifyCases([]caseSpec{{Name: "bad", Kind: "unknown"}}); err == nil {
		t.Fatal("verifyCases accepted unknown kind")
	}
}

func TestExpectedEventsAndNeedlesRejectDrift(t *testing.T) {
	t.Parallel()
	if err := verifyExpectedEvents([]string{"a"}, []string{"a", "b"}); err == nil {
		t.Fatal("verifyExpectedEvents accepted count drift")
	}
	if err := verifyExpectedEvents([]string{"a"}, []string{"b"}); err == nil {
		t.Fatal("verifyExpectedEvents accepted order drift")
	}
	check := sourceCheck{Name: "needle", File: "source.go", Contains: []string{"missing"}}
	if err := verifyNeedles(check, "present"); err == nil {
		t.Fatal("verifyNeedles accepted missing text")
	}
}

func TestVerifySourceChecksRejectsReadAndContentDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := verifySourceChecks(root, []sourceCheck{{Name: "missing", File: "missing.go"}}); err == nil {
		t.Fatal("verifySourceChecks accepted missing file")
	}
	path := filepath.Join(root, "source.go")
	if err := os.WriteFile(path, []byte("present"), 0o644); err != nil {
		t.Fatal(err)
	}
	check := sourceCheck{Name: "needle", File: "source.go", Contains: []string{"missing"}}
	if err := verifySourceChecks(root, []sourceCheck{check}); err == nil {
		t.Fatal("verifySourceChecks accepted missing needle")
	}
}

func TestCaseSummaryFallsBackForUnknownKind(t *testing.T) {
	t.Parallel()
	if got := caseSummary(caseSpec{Kind: "unknown"}); got != "unknown" {
		t.Fatalf("caseSummary unknown = %q", got)
	}
}
