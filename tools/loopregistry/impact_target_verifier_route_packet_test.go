package main

import "testing"

func TestTargetVerifierPlanExposesPathRoutingPackets(t *testing.T) {
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"tools/a.go", "tools/b.go"},
		AddedClaims:  []impactClaim{{ID: "claim-b"}},
	}
	attachTargetVerifierPlan(
		impact,
		routePacketIndex(),
		routePacketSurfaces(),
	)
	routes := impact.TargetVerifierPlan.RoutingPackets
	if len(routes) != 2 {
		t.Fatalf("routing packets = %+v", routes)
	}
	if impact.TargetVerifierPlan.RoutingPacketCount != 2 ||
		impact.TargetVerifierPlan.RoutingDirectCount != 1 ||
		impact.TargetVerifierPlan.RoutingFallbackCount != 1 {
		t.Fatalf("routing counts = %+v", impact.TargetVerifierPlan)
	}
	if routes[0].Path != "tools/a.go" ||
		!routes[0].UsesEntrypointFallback ||
		routes[0].DirectCommandCount != 0 ||
		routes[0].RunnableCommands[0] != "test-b" {
		t.Fatalf("route a = %+v", routes[0])
	}
	if routes[1].Path != "tools/b.go" ||
		routes[1].UsesEntrypointFallback ||
		routes[1].DirectCommandCount != 1 ||
		routes[1].RunnableCommands[0] != "test-b" {
		t.Fatalf("route b = %+v", routes[1])
	}
}

func routePacketIndex() architectureIndex {
	return architectureIndex{Paths: []architecturePathBinding{
		routePacketPath("tools/a.go", "claim-a", "test-a"),
		routePacketPath("tools/b.go", "claim-b", "test-b"),
	}}
}

func routePacketPath(path, claimID, command string) architecturePathBinding {
	return architecturePathBinding{
		Path:             path,
		Kind:             "code",
		LoopIDs:          []string{"loop-a"},
		ClaimIDs:         []string{claimID},
		VerifierCommands: []string{command},
		EvidenceChainIDs: []string{"chain-a"},
	}
}

func routePacketSurfaces() []claimSurface {
	return []claimSurface{
		{ID: "claim-a", VerifierCommands: []string{"test-a"}},
		{ID: "claim-b", VerifierCommands: []string{"test-b"}},
	}
}
