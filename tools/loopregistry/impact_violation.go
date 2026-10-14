package main

func impactViolationForClaim(
	scope string,
	claim claimBinding,
	record impactClaim,
	err error,
) impactViolation {
	return impactViolation{
		ClaimID:                   claim.ID,
		Scope:                     scope,
		Reason:                    err.Error(),
		RequiredBoundFiles:        claimRequiredImpactFiles(claim),
		RequiredEvidence:          claimEvidenceFiles(claim),
		RequiredReasoningEvidence: claimReasoningEvidenceFiles(),
		ChangedBoundFiles:         record.ChangedBoundFiles,
		ChangedEvidence:           record.ChangedEvidence,
		ChangedReasoningEvidence:  record.ChangedReasoningEvidence,
	}
}

func impactViolationForBoundSurface(
	claim claimBinding,
	record impactBoundSurface,
	err error,
) impactViolation {
	return impactViolation{
		ClaimID:                   claim.ID,
		Scope:                     "bound_surface",
		Reason:                    err.Error(),
		RequiredBoundFiles:        claimRequiredImpactFiles(claim),
		RequiredEvidence:          claimEvidenceFiles(claim),
		RequiredReasoningEvidence: claimReasoningEvidenceFiles(),
		ChangedBoundFiles:         record.ChangedBoundFiles,
		ChangedEvidence:           record.ChangedEvidence,
		ChangedReasoningEvidence:  record.ChangedReasoningEvidence,
	}
}
