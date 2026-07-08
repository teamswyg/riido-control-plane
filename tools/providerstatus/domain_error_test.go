package main

import (
	"testing"
)

func TestProviderStatusContainsRequireMeaningfulFields(t *testing.T) {
	if hasSurface([]surface{{Name: "ProviderStatusRecord"}}, "ProviderStatusRecord") {
		t.Fatal("surface without role must not satisfy requirement")
	}
	if hasValue([]value{{Value: "available"}}, "available") {
		t.Fatal("value without owner must not satisfy requirement")
	}
	if hasRule([]rule{{ID: "agent-id-required"}}, "agent-id-required") {
		t.Fatal("rule without text must not satisfy requirement")
	}
}

func TestProviderStatusVerifyDomainRejectsMissingVocabulary(t *testing.T) {
	if err := verifyDomain(minimalProviderStatusManifest()); err == nil {
		t.Fatal("expected missing surface error")
	}
	m := completeProviderStatusManifest()
	m.RoutingStatuses = nil
	if err := verifyDomain(m); err == nil {
		t.Fatal("expected missing routing status error")
	}
	m = completeProviderStatusManifest()
	m.DistributionChannels = nil
	if err := verifyDomain(m); err == nil {
		t.Fatal("expected missing distribution channel error")
	}
}

func TestProviderStatusVerifyDomainRejectsMissingRules(t *testing.T) {
	m := completeProviderStatusManifest()
	m.ValidationRules = nil
	if err := verifyDomain(m); err == nil {
		t.Fatal("expected missing validation rule error")
	}
	m = completeProviderStatusManifest()
	m.RoutingRules = nil
	if err := verifyDomain(m); err == nil {
		t.Fatal("expected missing routing rule error")
	}
}
