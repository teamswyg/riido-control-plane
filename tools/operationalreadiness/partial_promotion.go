package main

func partialPromotionFor(partials []partialCheck) partialPromotion {
	candidateIDs := []string{}
	staleIDs := []string{}
	for _, partial := range partials {
		if partial.Stale {
			candidateIDs = append(candidateIDs, partialCandidateID(partial))
			staleIDs = append(staleIDs, partial.ID)
		}
	}
	return partialPromotion{
		CandidateArtifact: readinessCandidateArtifact,
		CandidateCount:    len(candidateIDs),
		CandidateIDs:      candidateIDs,
		StalePartialCount: len(staleIDs),
		StalePartialIDs:   staleIDs,
		StaleAfterDays:    stalePartialAfterDays,
	}
}

func partialCandidateID(partial partialCheck) string {
	return readinessCandidateSourceID + ":" + partial.ID
}
