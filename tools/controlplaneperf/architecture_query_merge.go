package main

func mergeArchitectureQueryFallback(
	target *architectureFileEvidence,
	row architectureFileEvidence,
) {
	target.ComponentIDs = appendAllUnique(target.ComponentIDs, row.ComponentIDs)
	target.HotPathIDs = appendAllUnique(target.HotPathIDs, row.HotPathIDs)
	target.HotPathCategories = appendAllUnique(target.HotPathCategories, row.HotPathCategories)
	target.PressureDimensions = appendAllUnique(target.PressureDimensions, row.PressureDimensions)
	target.ObservabilitySignals = appendAllUnique(target.ObservabilitySignals, row.ObservabilitySignals)
	target.EvidenceRefs = appendAllUnique(target.EvidenceRefs, row.EvidenceRefs)
	target.TargetVerifierCommands = appendAllUnique(target.TargetVerifierCommands, row.TargetVerifierCommands)
	target.OptimizationCandidates = appendAllUnique(target.OptimizationCandidates, row.OptimizationCandidates)
}
