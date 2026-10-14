package main

import "testing"

func TestOperationalReadinessEvidenceIncludesInternalCompletion(t *testing.T) {
	got := newCompletionEvidence(loadManifestForTest(t).Checks)
	if got.Status != "verified" || got.ThresholdBasisPoints != 9000 {
		t.Fatalf("unexpected completion gate: %+v", got)
	}
	if got.InternalCheckCount != 5 || got.InternalCoveredCount != 5 || got.InternalPartialCount != 0 {
		t.Fatalf("unexpected internal counts: %+v", got)
	}
	if got.ExternalExcludedCount != 9 || got.ExternalPartialCount != 9 {
		t.Fatalf("unexpected external counts: %+v", got)
	}
	if got.InternalCompletenessBasisPoints != 10000 {
		t.Fatalf("internal completeness = %d", got.InternalCompletenessBasisPoints)
	}
}

func TestOperationalReadinessRejectsIncompleteInternalScope(t *testing.T) {
	m := loadManifestForTest(t)
	for i := range m.Checks {
		if m.Checks[i].CompletionScope == completionScopeInternal {
			m.Checks[i].Status = "partial"
			break
		}
	}
	if err := verifyInternalCompletion(m.Checks); err == nil {
		t.Fatal("expected internal completion below threshold to fail")
	}
}

func TestOperationalReadinessRejectsMissingCompletionScopeReason(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].ScopeReason = ""
	if err := verifyCompletionScope(m.Checks[0]); err == nil {
		t.Fatal("expected missing completion scope reason to fail")
	}
}
