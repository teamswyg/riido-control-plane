package main

import "testing"

func TestHarnessProviderCoverageRequiresDeclaredProviders(t *testing.T) {
	err := verifyHarnessProviderCoverage(loopRecord{
		ID:       "provider_acceptance_harness",
		Kind:     kindHarness,
		Observes: []string{"codex_provider_qa", "openclaw_provider_qa"},
	})
	if err == nil {
		t.Fatal("expected missing provider coverage to fail")
	}
}

func TestHarnessProviderCoverageRejectsProviderDrift(t *testing.T) {
	err := verifyHarnessProviderCoverage(loopRecord{
		ID:        "provider_acceptance_harness",
		Kind:      kindHarness,
		Observes:  []string{"codex_provider_qa"},
		Providers: []string{"codex", "openclaw"},
	})
	if err == nil {
		t.Fatal("expected unobserved provider coverage to fail")
	}
}

func TestHarnessProviderCoverageAcceptsMatchingProviders(t *testing.T) {
	err := verifyHarnessProviderCoverage(loopRecord{
		ID:        "provider_acceptance_harness",
		Kind:      kindHarness,
		Observes:  []string{"codex_provider_qa", "openclaw_provider_qa", "daemon_runtime_snapshot"},
		Providers: []string{"openclaw", "codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderCoverageEvidenceSortsProviders(t *testing.T) {
	got := providerCoverage([]loopRecord{{
		ID:        "provider_acceptance_harness",
		Providers: []string{"openclaw", "codex"},
	}})
	providers := got["provider_acceptance_harness"]
	if len(providers) != 2 || providers[0] != "codex" || providers[1] != "openclaw" {
		t.Fatalf("providers = %+v", providers)
	}
}
