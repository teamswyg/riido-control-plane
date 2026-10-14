package main

import "testing"

func TestImpactTargetVerifierPlanUsesArchitectureIndex(t *testing.T) {
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"internal/unbound.go", "tools/example/a.go"},
	}
	attachTargetVerifierPlan(impact, architectureIndex{
		Paths: []architecturePathBinding{{
			Path:             "tools/example/a.go",
			Kind:             "code",
			LoopIDs:          []string{"loop-a"},
			ClaimIDs:         []string{"claim-a"},
			VerifierCommands: []string{"go test ./tools/example -count=1"},
			EvidenceChainIDs: []string{"chain-a"},
		}},
	}, nil)
	plan := impact.TargetVerifierPlan
	if plan == nil || plan.ChangedPathCount != 2 ||
		plan.MatchedPathCount != 1 || plan.ComponentCount != 1 ||
		plan.CommandCount != 1 || plan.ExactPathCount != 1 ||
		plan.ComponentRouteCount != 0 {
		t.Fatalf("target verifier plan = %+v", plan)
	}
	if plan.VerifierCommands[0] != "go test ./tools/example -count=1" {
		t.Fatalf("plan verifier commands = %#v", plan.VerifierCommands)
	}
	if len(plan.CommandUnits) != 1 ||
		plan.CommandUnits[0].PathCount != 1 ||
		plan.CommandUnits[0].ClaimIDs[0] != "claim-a" {
		t.Fatalf("plan command units = %+v", plan.CommandUnits)
	}
	if len(plan.EntrypointCommands) != 1 ||
		plan.EntrypointCommands[0] != "go test ./tools/example -count=1" {
		t.Fatalf("plan entrypoint commands = %+v", plan.EntrypointCommands)
	}
	path := plan.Paths[0]
	if path.Path != "tools/example/a.go" ||
		path.Component != "tools/example" ||
		path.MatchKind != "exact" ||
		path.ClaimIDs[0] != "claim-a" ||
		path.VerifierCommands[0] != "go test ./tools/example -count=1" {
		t.Fatalf("target verifier path = %+v", path)
	}
	component := plan.Components[0]
	if component.Component != "tools/example" ||
		component.PathCount != 1 ||
		component.EvidenceChainIDs[0] != "chain-a" {
		t.Fatalf("target verifier component = %+v", component)
	}
}
