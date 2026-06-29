package main

import "testing"

func TestChainImpactRequiresExecutableRefChange(t *testing.T) {
	base := []chain{testChain("chain", "old")}
	current := []chain{testChain("chain", "new")}
	if _, err := verifyChainImpact("origin/main", base, current,
		map[string]bool{defaultManifest: true}); err == nil {
		t.Fatal("expected chain change without executable ref change to fail")
	}
}

func TestChainImpactAllowsExecutableRefChange(t *testing.T) {
	base := []chain{testChain("chain", "old")}
	current := []chain{testChain("chain", "new")}
	evidence, err := verifyChainImpact("origin/main", base, current,
		map[string]bool{"tools/evidencegraph/run.go": true, defaultManifest: true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.ChangedChainCount != 1 || len(evidence.ChangedChains[0].ChangedExecutableRefs) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
	chain := evidence.ChangedChains[0]
	if chain.Claims[0] != "claim" ||
		chain.VerifierRefs[0] != "tools/evidencegraph/impact_test.go" ||
		chain.EvidenceRefs[0] != "evidence-graph-evidence" ||
		chain.NextLoop != "loop" {
		t.Fatalf("chain scope = %+v", chain)
	}
}

func testChain(id, observation string) chain {
	return chain{
		ID: id, Observation: observation, Hypothesis: "hypothesis", Decision: "decision", NextLoop: "loop",
		Claims:    []string{"claim"},
		Changes:   []ref{{Kind: "tool", Path: "tools/evidencegraph/run.go"}},
		Verifiers: []ref{{Kind: "test", Path: "tools/evidencegraph/impact_test.go"}},
		Evidence:  []ref{{Kind: "artifact", Path: "evidence-graph-evidence", Redacted: true}},
	}
}
