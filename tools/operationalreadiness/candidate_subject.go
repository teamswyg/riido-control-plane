package main

func partialSubject(partial partialCheck) *candidateSubject {
	return &candidateSubject{
		Kind:           "operational_readiness_partial",
		CheckID:        partial.ID,
		Category:       partial.Category,
		AgeDays:        partial.AgeDays,
		Stale:          partial.Stale,
		NextArtifact:   partial.NextArtifact,
		NextCommand:    partial.NextCommand,
		StaleAfterDays: stalePartialAfterDays,
	}
}
