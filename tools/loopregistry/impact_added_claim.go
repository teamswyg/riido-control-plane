package main

import "fmt"

func verifyAddedClaimImpact(claim claimBinding, changed map[string]bool) (impactClaim, error) {
	record := impactClaim{ID: claim.ID}
	for _, path := range claimRequiredImpactFiles(claim) {
		if changed[path] {
			record.ChangedBoundFiles = append(record.ChangedBoundFiles, path)
		}
	}
	if len(record.ChangedBoundFiles) == 0 {
		return record, fmt.Errorf("claim %s added without a bound code or test file change", claim.ID)
	}
	record.ChangedEvidence = changedValues(claimEvidenceFiles(claim), changed)
	if len(record.ChangedEvidence) == 0 {
		return record, fmt.Errorf("claim %s added without generated claim evidence change", claim.ID)
	}
	record.ChangedReasoningEvidence = changedValues(claimReasoningEvidenceFiles(), changed)
	if len(record.ChangedReasoningEvidence) == 0 {
		return record, fmt.Errorf("claim %s added without evidence graph reasoning change", claim.ID)
	}
	return record, nil
}
