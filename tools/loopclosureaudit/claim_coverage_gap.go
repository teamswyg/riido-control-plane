package main

import "sort"

func claimCoverageGaps(deps dependencies) []claimCoverageGap {
	loops := registryLoopsByID(deps.registry.Loops)
	gaps := []claimCoverageGap{}
	for _, claim := range deps.registry.Claims {
		loop, ok := loops[claim.Loop]
		if !ok {
			continue
		}
		missing := missingCoverageDimensions(claim, loop)
		if len(missing) == 0 {
			continue
		}
		gaps = append(gaps, claimCoverageGap{
			ClaimID:           claim.ID,
			Loop:              claim.Loop,
			MissingDimensions: missing,
		})
	}
	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Loop != gaps[j].Loop {
			return gaps[i].Loop < gaps[j].Loop
		}
		return gaps[i].ClaimID < gaps[j].ClaimID
	})
	return gaps
}

func registryLoopsByID(loops []registryLoop) map[string]registryLoop {
	out := map[string]registryLoop{}
	for _, loop := range loops {
		out[loop.ID] = loop
	}
	return out
}
