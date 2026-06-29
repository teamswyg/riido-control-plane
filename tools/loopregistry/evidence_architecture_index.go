package main

func architectureIndexFor(
	claims []claimBinding,
	surfaces []claimSurface,
) architectureIndex {
	loops := claimLoops(claims)
	byPath := map[string]*architecturePathBinding{}
	for _, surface := range surfaces {
		loopID := loops[surface.ID]
		addArchitecturePaths(byPath, surface.CodePaths, "code", surface, loopID)
		addArchitecturePaths(byPath, surface.TestPaths, "test", surface, loopID)
		addArchitecturePaths(byPath, surface.ManifestPaths, "manifest", surface, loopID)
		addArchitecturePaths(byPath, surface.GeneratedDocs, "generated_doc", surface, loopID)
	}
	out := architectureIndex{Paths: architecturePathBindings(byPath)}
	out.Components = architectureComponents(out.Paths)
	out.PathCount = len(out.Paths)
	out.ComponentCount = len(out.Components)
	for _, binding := range out.Paths {
		out.BindingCount += len(binding.ClaimIDs)
		out.VerifierCommandCount += len(binding.VerifierCommands)
	}
	return out
}

func claimLoops(claims []claimBinding) map[string]string {
	out := map[string]string{}
	for _, claim := range claims {
		out[claim.ID] = claim.Loop
	}
	return out
}

func addArchitecturePaths(
	byPath map[string]*architecturePathBinding,
	paths []string,
	kind string,
	surface claimSurface,
	loopID string,
) {
	for _, path := range paths {
		binding := architecturePath(byPath, path, kind)
		binding.LoopIDs = appendUnique(binding.LoopIDs, loopID)
		binding.ClaimIDs = appendUnique(binding.ClaimIDs, surface.ID)
		binding.VerifierCommands = appendUnique(binding.VerifierCommands, surface.VerifierCommands...)
		binding.EvidenceChainIDs = appendUnique(binding.EvidenceChainIDs, surface.EvidenceChainIDs...)
	}
}

func architecturePath(
	byPath map[string]*architecturePathBinding,
	path string,
	kind string,
) *architecturePathBinding {
	if byPath[path] == nil {
		byPath[path] = &architecturePathBinding{Path: path, Kind: kind}
		return byPath[path]
	}
	if byPath[path].Kind != kind {
		byPath[path].Kind = "mixed"
	}
	return byPath[path]
}
