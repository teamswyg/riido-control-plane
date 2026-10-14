package main

import "testing"

func TestRunCaseRejectsUnknownKind(t *testing.T) {
	if _, err := runCase(caseSpec{Name: "unknown", Kind: "wat"}); err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestVerifyCasesReturnsCaseError(t *testing.T) {
	cases := []caseSpec{{Name: "unknown", Kind: "wat"}}
	if _, err := verifyCases(cases); err == nil {
		t.Fatal("expected case error")
	}
}

func TestVerifyCaseNamesRejectsCountAndMissingName(t *testing.T) {
	cases := []caseSpec{{Name: "one"}, {Name: "two"}}
	if err := verifyCaseNames(cases, nil); err == nil {
		t.Fatal("expected count error")
	}
	results := []caseEvidence{{Name: "one"}, {Name: "other"}}
	if err := verifyCaseNames(cases, results); err == nil {
		t.Fatal("expected missing case evidence error")
	}
}

func TestReviewSeedCaseVerifiersRejectMismatchedExpectations(t *testing.T) {
	if _, err := verifyProvisionCase(caseSpec{
		Name: "provision", Kind: "provision", WantRawTokenPresent: true,
	}); err == nil {
		t.Fatal("expected provision mismatch")
	}
	if _, err := verifyCatalogCase(caseSpec{
		Name: "catalog", Kind: "catalog", WantVisibleAgents: []string{"wrong"},
	}); err == nil {
		t.Fatal("expected catalog mismatch")
	}
	if _, err := verifyProviderStatusCase(caseSpec{
		Name: "provider", Kind: "provider-status", WantProviderCount: -1,
	}); err == nil {
		t.Fatal("expected provider status mismatch")
	}
	if _, err := verifyHTTPCase(caseSpec{
		Name: "http", Kind: "http", WantCatalogStatus: 201,
	}); err == nil {
		t.Fatal("expected http status mismatch")
	}
}

func TestCaseSummaryFallsBackForUnknownKind(t *testing.T) {
	if got := caseSummary(caseSpec{Kind: "wat"}); got != "unknown" {
		t.Fatalf("summary = %q", got)
	}
}
