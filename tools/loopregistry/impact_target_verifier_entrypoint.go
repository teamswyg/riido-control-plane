package main

import "sort"

const targetVerifierEntrypointLimit = 5

func targetVerifierEntrypointCommands(
	units []targetVerifierCommand,
) []string {
	if len(units) == 0 {
		return nil
	}
	ranked := append([]targetVerifierCommand(nil), units...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return targetVerifierCommandLess(ranked[i], ranked[j])
	})
	limit := targetVerifierEntrypointLimit
	if len(ranked) < limit {
		limit = len(ranked)
	}
	out := make([]string, 0, limit)
	for _, unit := range ranked[:limit] {
		out = append(out, unit.Command)
	}
	return out
}

func targetVerifierCommandLess(
	left, right targetVerifierCommand,
) bool {
	if left.PathCount != right.PathCount {
		return left.PathCount > right.PathCount
	}
	if left.ComponentCount != right.ComponentCount {
		return left.ComponentCount > right.ComponentCount
	}
	if len(left.ClaimIDs) != len(right.ClaimIDs) {
		return len(left.ClaimIDs) > len(right.ClaimIDs)
	}
	if len(left.EvidenceChainIDs) != len(right.EvidenceChainIDs) {
		return len(left.EvidenceChainIDs) > len(right.EvidenceChainIDs)
	}
	return left.Command < right.Command
}
