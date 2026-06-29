package main

func attachTargetVerifierPlan(
	impact *impactEvidence,
	index architectureIndex,
) {
	if impact == nil || !impact.Enabled {
		return
	}
	plan := targetVerifierPlan{
		ChangedPathCount: len(impact.ChangedFiles),
		Paths:            targetVerifierPaths(impact.ChangedFiles, index),
	}
	plan.MatchedPathCount = len(plan.Paths)
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
	plan.EntrypointCommands = targetVerifierEntrypointCommands(plan.CommandUnits)
	impact.TargetVerifierPlan = &plan
}

func targetVerifierPaths(
	changedFiles []string,
	index architectureIndex,
) []targetVerifierPath {
	byPath := architecturePathsByPath(index.Paths)
	out := []targetVerifierPath{}
	for _, path := range sortedCopy(changedFiles) {
		binding, ok := byPath[path]
		if !ok {
			continue
		}
		out = append(out, targetVerifierPath{
			Path:             binding.Path,
			Component:        architectureComponentID(binding.Path),
			Kind:             binding.Kind,
			LoopIDs:          sortedCopy(binding.LoopIDs),
			ClaimIDs:         sortedCopy(binding.ClaimIDs),
			VerifierCommands: sortedCopy(binding.VerifierCommands),
			EvidenceChainIDs: sortedCopy(binding.EvidenceChainIDs),
		})
	}
	return out
}

func architecturePathsByPath(
	bindings []architecturePathBinding,
) map[string]architecturePathBinding {
	out := map[string]architecturePathBinding{}
	for _, binding := range bindings {
		out[binding.Path] = binding
	}
	return out
}
