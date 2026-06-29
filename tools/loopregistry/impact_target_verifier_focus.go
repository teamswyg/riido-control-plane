package main

import "sort"

func attachFocusedTargetVerifierPlan(
	plan *targetVerifierPlan,
	impact *impactEvidence,
	surfaces []claimSurface,
) {
	if plan == nil || impact == nil {
		return
	}
	claimIDs := focusedTargetVerifierClaimIDs(impact)
	if len(claimIDs) == 0 {
		return
	}
	commands := focusedVerifierCommands(claimIDs, surfaces, plan.CommandUnits)
	plan.FocusedClaimIDs = claimIDs
	plan.FocusedCommandCount = len(commands)
	plan.FocusedCommands = commands
}

func focusedTargetVerifierClaimIDs(impact *impactEvidence) []string {
	ids := []string{}
	ids = appendImpactClaimIDs(ids, impact.AddedClaims)
	ids = appendImpactClaimIDs(ids, impact.Claims)
	ids = appendImpactClaimIDs(ids, impact.RemovedClaims)
	if len(ids) == 0 {
		ids = appendBoundSurfaceIDs(ids, impact.BoundSurfaces)
	}
	sort.Strings(ids)
	return ids
}

func appendImpactClaimIDs(ids []string, claims []impactClaim) []string {
	for _, claim := range claims {
		ids = appendUnique(ids, claim.ID)
	}
	return ids
}

func appendBoundSurfaceIDs(
	ids []string,
	surfaces []impactBoundSurface,
) []string {
	for _, surface := range surfaces {
		ids = appendUnique(ids, surface.ID)
	}
	return ids
}
