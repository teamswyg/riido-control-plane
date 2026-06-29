package main

import "testing"

func TestTargetVerifierCommandsGroupCommandImpact(t *testing.T) {
	paths := []targetVerifierPath{
		{
			Path:             "tools/example/a.go",
			Component:        "tools/example",
			LoopIDs:          []string{"loop-a"},
			ClaimIDs:         []string{"claim-a"},
			VerifierCommands: []string{"go test ./tools/example -count=1"},
			EvidenceChainIDs: []string{"chain-a"},
		},
		{
			Path:             "tools/example/b.go",
			Component:        "tools/example",
			LoopIDs:          []string{"loop-b"},
			ClaimIDs:         []string{"claim-b"},
			VerifierCommands: []string{"go test ./tools/example -count=1"},
			EvidenceChainIDs: []string{"chain-b"},
		},
	}
	units := targetVerifierCommands(paths)
	if len(units) != 1 {
		t.Fatalf("units = %+v", units)
	}
	unit := units[0]
	if unit.Command != "go test ./tools/example -count=1" ||
		unit.PathCount != 2 ||
		unit.ComponentCount != 1 ||
		len(unit.ClaimIDs) != 2 ||
		len(unit.EvidenceChainIDs) != 2 {
		t.Fatalf("unit = %+v", unit)
	}
}
