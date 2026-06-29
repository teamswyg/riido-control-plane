package main

import "sort"

func targetVerifierComponents(
	paths []targetVerifierPath,
) []targetVerifierComponent {
	byComponent := map[string]*targetVerifierComponent{}
	for _, path := range paths {
		component := targetVerifierComponentFor(byComponent, path.Component)
		component.PathCount++
		component.LoopIDs = appendUnique(component.LoopIDs, path.LoopIDs...)
		component.ClaimIDs = appendUnique(component.ClaimIDs, path.ClaimIDs...)
		component.VerifierCommands = appendUnique(component.VerifierCommands, path.VerifierCommands...)
		component.EvidenceChainIDs = appendUnique(component.EvidenceChainIDs, path.EvidenceChainIDs...)
	}
	names := targetVerifierComponentNames(byComponent)
	out := make([]targetVerifierComponent, 0, len(names))
	for _, name := range names {
		out = append(out, *byComponent[name])
	}
	return out
}

func targetVerifierComponentNames(
	byComponent map[string]*targetVerifierComponent,
) []string {
	names := make([]string, 0, len(byComponent))
	for name := range byComponent {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func targetVerifierComponentFor(
	byComponent map[string]*targetVerifierComponent,
	name string,
) *targetVerifierComponent {
	if byComponent[name] == nil {
		byComponent[name] = &targetVerifierComponent{Component: name}
	}
	return byComponent[name]
}
