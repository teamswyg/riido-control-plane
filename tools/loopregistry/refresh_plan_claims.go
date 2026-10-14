package main

type refreshPlanClaimSet struct {
	ClaimIDs         []string
	EvidenceChainIDs []string
	VerifierCommands []string
	CommandScopes    map[string]refreshCommandScope
}

func refreshPlanClaimCoverage(
	claims []claimBinding,
	surfaces []claimSurface,
) map[string]refreshPlanClaimSet {
	byClaim := claimSurfacesByID(surfaces)
	out := map[string]refreshPlanClaimSet{}
	for _, claim := range claims {
		surface := byClaim[claim.ID]
		set := out[claim.Loop]
		set.CommandScopes = ensureRefreshCommandScopes(set.CommandScopes)
		set.ClaimIDs = append(set.ClaimIDs, claim.ID)
		set.VerifierCommands = appendMissingStrings(
			set.VerifierCommands,
			surface.VerifierCommands,
		)
		set.EvidenceChainIDs = appendMissingStrings(set.EvidenceChainIDs, surface.EvidenceChainIDs)
		addRefreshCommandScopes(set.CommandScopes, claim.ID, surface)
		out[claim.Loop] = set
	}
	return out
}

func claimSurfacesByID(surfaces []claimSurface) map[string]claimSurface {
	out := map[string]claimSurface{}
	for _, surface := range surfaces {
		out[surface.ID] = surface
	}
	return out
}

func appendMissingStrings(values, candidates []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, candidate := range candidates {
		if !seen[candidate] {
			values = append(values, candidate)
			seen[candidate] = true
		}
	}
	return values
}
