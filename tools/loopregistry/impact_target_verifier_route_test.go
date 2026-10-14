package main

import "testing"

func TestImpactTargetVerifierPlanUsesComponentRouteForNewFiles(t *testing.T) {
	index := architectureIndexFor(
		[]claimBinding{{ID: "claim-a", Loop: "loop-a"}},
		[]claimSurface{{
			ID:               "claim-a",
			CodePaths:        []string{"tools/example/a.go"},
			VerifierCommands: []string{"go test ./tools/example -count=1"},
			EvidenceChainIDs: []string{"chain-a"},
		}},
	)
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"tools/example/new_file.go"},
	}
	attachTargetVerifierPlan(impact, index, nil)
	plan := impact.TargetVerifierPlan
	if plan == nil || plan.MatchedPathCount != 1 ||
		plan.ExactPathCount != 0 || plan.ComponentRouteCount != 1 {
		t.Fatalf("target verifier plan = %+v", plan)
	}
	path := plan.Paths[0]
	if path.Path != "tools/example/new_file.go" ||
		path.Component != "tools/example" ||
		path.Kind != "code" ||
		path.MatchKind != "component_route" ||
		path.ClaimIDs[0] != "claim-a" ||
		path.VerifierCommands[0] != "go test ./tools/example -count=1" {
		t.Fatalf("component-routed target verifier path = %+v", path)
	}
}
