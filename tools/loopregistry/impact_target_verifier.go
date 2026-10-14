package main

func attachTargetVerifierPlan(
	impact *impactEvidence,
	index architectureIndex,
	surfaces []claimSurface,
) {
	if impact == nil || !impact.Enabled {
		return
	}
	plan := targetVerifierPlan{
		ChangedPathCount: len(impact.ChangedFiles),
		Paths:            targetVerifierPaths(impact.ChangedFiles, index),
	}
	plan.MatchedPathCount = len(plan.Paths)
	plan.ExactPathCount, plan.ComponentRouteCount = targetVerifierPathMatchCounts(plan.Paths)
	plan.Components = targetVerifierComponents(plan.Paths)
	plan.ComponentCount = len(plan.Components)
	for _, path := range plan.Paths {
		plan.VerifierCommands = appendUnique(
			plan.VerifierCommands,
			path.VerifierCommands...,
		)
	}
	plan.CommandCount = len(plan.VerifierCommands)
	plan.CommandUnits = targetVerifierCommands(plan.Paths)
	attachFocusedTargetVerifierPlan(&plan, impact, surfaces)
	plan.EntrypointCommands = targetVerifierEntrypointCommands(plan.CommandUnits)
	attachRunnableTargetVerifierCommands(&plan)
	attachTargetVerifierRoutingPackets(&plan)
	impact.TargetVerifierPlan = &plan
}
