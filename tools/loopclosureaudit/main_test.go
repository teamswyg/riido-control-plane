package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoopClosureAuditManifestVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.RequirementCount < 4 || got.CheckCount < 16 || got.Status != "verified" {
		t.Fatalf("evidence coverage = %+v", got)
	}
}

func TestLoopClosureAuditRejectsMissingClaim(t *testing.T) {
	m, deps := loadForTest(t)
	deleteClaim(&deps.registry.Claims, "pre_commit_must_run_claim_binding_impact")
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected missing claim to fail")
	}
}

func TestLoopClosureAuditRejectsMissingHarnessPromotionTarget(t *testing.T) {
	m, deps := loadForTest(t)
	for i := range deps.registry.Loops {
		if deps.registry.Loops[i].ID == "provider_acceptance_harness" {
			deps.registry.Loops[i].PromotesTo = nil
		}
	}
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected missing harness promotion target to fail")
	}
}

func loadForTest(t *testing.T) (manifest, dependencies) {
	t.Helper()
	m, deps, err := loadAll("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	return m, deps
}

func readEvidence(t *testing.T, path string) evidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func deleteClaim(claims *[]registryClaim, id string) {
	for i, claim := range *claims {
		if claim.ID == id {
			*claims = append((*claims)[:i], (*claims)[i+1:]...)
			return
		}
	}
}
