package main

import (
	"strings"
	"testing"
)

func TestMaybeVerifyImpactSkipsEmptyBaseRef(t *testing.T) {
	t.Parallel()
	got, err := maybeVerifyImpact(".", defaultManifest, "", manifest{})
	if err != nil {
		t.Fatalf("maybeVerifyImpact: %v", err)
	}
	if got != nil {
		t.Fatalf("impact should be nil: %+v", got)
	}
}

func TestMaybeVerifyImpactReportsBaseLoadFailure(t *testing.T) {
	t.Parallel()
	_, err := maybeVerifyImpact(".", defaultManifest, "missing-ref", manifest{})
	if err == nil || !strings.Contains(err.Error(), "load base evidence graph") {
		t.Fatalf("error = %v, want base load failure", err)
	}
}

func TestChainImpactCoversAddedAndRemovedChains(t *testing.T) {
	t.Parallel()
	added := testChain("added", "added")
	removed := testChain("removed", "removed")
	changed := map[string]bool{"tools/evidencegraph/run.go": true}
	evidence, err := verifyChainImpact("HEAD", []chain{removed}, []chain{added}, changed)
	if err != nil {
		t.Fatalf("verifyChainImpact: %v", err)
	}
	if evidence.AddedChainCount != 1 || evidence.RemovedChainCount != 1 {
		t.Fatalf("impact evidence = %+v", evidence)
	}
}

func TestRemovedChainRequiresExecutableRefChange(t *testing.T) {
	t.Parallel()
	_, err := verifyChainImpact("HEAD", []chain{testChain("removed", "old")}, nil, map[string]bool{})
	if err == nil || !strings.Contains(err.Error(), "removed without executable ref change") {
		t.Fatalf("error = %v, want removed executable-ref failure", err)
	}
}
