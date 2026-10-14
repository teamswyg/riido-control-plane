package main

import "testing"

func TestTargetVerifierEntrypointsRankBroadCommands(t *testing.T) {
	got := targetVerifierEntrypointCommands([]targetVerifierCommand{
		{
			Command:        "go test ./tools/small -count=1",
			PathCount:      1,
			ComponentCount: 1,
			ClaimIDs:       []string{"claim-a"},
		},
		{
			Command:          "go test ./... -count=1",
			PathCount:        3,
			ComponentCount:   2,
			ClaimIDs:         []string{"claim-a", "claim-b"},
			EvidenceChainIDs: []string{"chain-a", "chain-b"},
		},
	})
	if len(got) != 2 || got[0] != "go test ./... -count=1" {
		t.Fatalf("entrypoints = %#v", got)
	}
}

func TestTargetVerifierEntrypointsLimitOutput(t *testing.T) {
	units := []targetVerifierCommand{}
	for _, command := range []string{"a", "b", "c", "d", "e", "f"} {
		units = append(units, targetVerifierCommand{
			Command:   command,
			PathCount: 1,
		})
	}
	got := targetVerifierEntrypointCommands(units)
	if len(got) != targetVerifierEntrypointLimit {
		t.Fatalf("entrypoints = %#v", got)
	}
}
