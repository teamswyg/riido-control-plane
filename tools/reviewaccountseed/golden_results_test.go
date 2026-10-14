package main

import "testing"

func assertReviewAccountSeedResults(t *testing.T, got []caseEvidence) {
	t.Helper()
	want := []caseEvidence{
		{Name: "credential-redacted", Kind: "provision", TokenHashPresent: true},
		{Name: "catalog-visibility", Kind: "catalog", VisibleAgents: []string{"review-other-public", "review-owned-private", "review-owned-public"}},
		{Name: "provider-status-non-routable", Kind: "provider-status", ProviderCount: 4, Channel: "mac-app-store"},
		{Name: "http-review-read-deny-poll", Kind: "http", CatalogStatus: 200, ProviderStatus: 200, PollStatus: 403},
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected result count: %+v", got)
	}
	for i := range want {
		assertReviewAccountSeedResult(t, i, got[i], want[i])
	}
}

func assertReviewAccountSeedResult(t *testing.T, i int, got, want caseEvidence) {
	t.Helper()
	if got.Name != want.Name || got.Kind != want.Kind || got.RawTokenPresent {
		t.Fatalf("unexpected result[%d]: %+v", i, got)
	}
	if got.TokenHashPresent != want.TokenHashPresent || got.ProviderCount != want.ProviderCount {
		t.Fatalf("unexpected result[%d] counters: %+v", i, got)
	}
	if got.Channel != want.Channel || got.CatalogStatus != want.CatalogStatus || got.PollStatus != want.PollStatus {
		t.Fatalf("unexpected result[%d] status: %+v", i, got)
	}
}
