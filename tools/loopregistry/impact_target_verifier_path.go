package main

func targetVerifierPaths(
	changedFiles []string,
	index architectureIndex,
) []targetVerifierPath {
	byPath := architecturePathsByPath(index.Paths)
	byComponent := architectureComponentsByName(index.Components)
	out := []targetVerifierPath{}
	for _, path := range sortedCopy(changedFiles) {
		if binding, ok := byPath[path]; ok {
			out = append(out, targetVerifierPathFromBinding(binding))
			continue
		}
		if route, ok := byComponent[architectureComponentID(path)]; ok {
			out = append(out, targetVerifierPathFromComponent(path, route))
		}
	}
	return out
}

func targetVerifierPathFromBinding(
	binding architecturePathBinding,
) targetVerifierPath {
	return targetVerifierPath{
		Path:             binding.Path,
		Component:        architectureComponentID(binding.Path),
		Kind:             binding.Kind,
		MatchKind:        "exact",
		LoopIDs:          sortedCopy(binding.LoopIDs),
		ClaimIDs:         sortedCopy(binding.ClaimIDs),
		VerifierCommands: sortedCopy(binding.VerifierCommands),
		EvidenceChainIDs: sortedCopy(binding.EvidenceChainIDs),
	}
}

func targetVerifierPathFromComponent(
	path string,
	route architectureComponent,
) targetVerifierPath {
	return targetVerifierPath{
		Path:             path,
		Component:        route.Component,
		Kind:             targetVerifierPathKind(path),
		MatchKind:        "component_route",
		LoopIDs:          sortedCopy(route.LoopIDs),
		ClaimIDs:         sortedCopy(route.ClaimIDs),
		VerifierCommands: sortedCopy(route.VerifierCommands),
		EvidenceChainIDs: sortedCopy(route.EvidenceChainIDs),
	}
}

func targetVerifierPathKind(path string) string {
	if isClaimTestPath(path) {
		return "test"
	}
	if isClaimManifestPath(path) {
		return "manifest"
	}
	return "code"
}
