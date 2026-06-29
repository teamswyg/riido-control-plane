package main

import "sort"

func architectureComponents(paths []architecturePathBinding) []architectureComponent {
	byComponent := map[string]*architectureComponent{}
	for _, path := range paths {
		component := architectureComponentFor(byComponent, architectureComponentID(path.Path))
		component.PathCount++
		component.LoopIDs = appendUnique(component.LoopIDs, path.LoopIDs...)
		component.ClaimIDs = appendUnique(component.ClaimIDs, path.ClaimIDs...)
		component.VerifierCommands = appendUnique(component.VerifierCommands, path.VerifierCommands...)
		component.EvidenceChainIDs = appendUnique(component.EvidenceChainIDs, path.EvidenceChainIDs...)
	}
	names := make([]string, 0, len(byComponent))
	for name := range byComponent {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]architectureComponent, 0, len(names))
	for _, name := range names {
		out = append(out, *byComponent[name])
	}
	return out
}

func architectureComponentFor(
	byComponent map[string]*architectureComponent,
	name string,
) *architectureComponent {
	if byComponent[name] == nil {
		byComponent[name] = &architectureComponent{Component: name}
	}
	return byComponent[name]
}
