package main

import "testing"

func TestProofKeyIdentifiesExecutableSurfaces(t *testing.T) {
	cases := map[string]check{
		"claim:claim-a": {
			Kind: "claim", ID: "claim-a",
		},
		"workflow:.github/workflows/a.yml:-impact-base,evidence": {
			Kind:     "workflow",
			Path:     ".github/workflows/a.yml",
			Contains: []string{"-impact-base", "evidence"},
		},
		"graph_edge:loop:enforces:claim": {
			Kind: "graph_edge", From: "loop", Relation: "enforces", To: "claim",
		},
	}
	for want, check := range cases {
		if got := proofKey(check); got != want {
			t.Fatalf("proofKey(%+v) = %q, want %q", check, got, want)
		}
	}
}

func TestRequirementProofsAreVerified(t *testing.T) {
	got := requirementProofs([]check{{Kind: "loop", ID: "loop-a"}})
	if len(got) != 1 || got[0].Status != "verified" ||
		got[0].Key != "loop:loop-a" {
		t.Fatalf("proofs = %+v", got)
	}
}
