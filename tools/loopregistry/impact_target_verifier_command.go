package main

import "sort"

func targetVerifierCommands(
	paths []targetVerifierPath,
) []targetVerifierCommand {
	byCommand := map[string]*targetVerifierCommand{}
	for _, path := range paths {
		for _, command := range path.VerifierCommands {
			unit := targetVerifierCommandFor(byCommand, command)
			unit.Paths = appendUnique(unit.Paths, path.Path)
			unit.Components = appendUnique(unit.Components, path.Component)
			unit.LoopIDs = appendUnique(unit.LoopIDs, path.LoopIDs...)
			unit.ClaimIDs = appendUnique(unit.ClaimIDs, path.ClaimIDs...)
			unit.EvidenceChainIDs = appendUnique(
				unit.EvidenceChainIDs,
				path.EvidenceChainIDs...,
			)
		}
	}
	names := targetVerifierCommandNames(byCommand)
	out := make([]targetVerifierCommand, 0, len(names))
	for _, name := range names {
		unit := *byCommand[name]
		unit.PathCount = len(unit.Paths)
		unit.ComponentCount = len(unit.Components)
		out = append(out, unit)
	}
	return out
}

func targetVerifierCommandNames(
	byCommand map[string]*targetVerifierCommand,
) []string {
	names := make([]string, 0, len(byCommand))
	for name := range byCommand {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func targetVerifierCommandFor(
	byCommand map[string]*targetVerifierCommand,
	command string,
) *targetVerifierCommand {
	if byCommand[command] == nil {
		byCommand[command] = &targetVerifierCommand{Command: command}
	}
	return byCommand[command]
}
