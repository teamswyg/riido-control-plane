package main

import (
	"fmt"
	"strings"
)

func claimsByID(claims []claimBinding) map[string]claimBinding {
	out := map[string]claimBinding{}
	for _, claim := range claims {
		out[claim.ID] = claim
	}
	return out
}

func claimSignature(claim claimBinding) string {
	parts := []string{claim.ID, claim.Statement, claim.Loop}
	parts = append(parts, prefixedValues("file", claim.Files)...)
	parts = append(parts, prefixedValues("verifier", claim.Verifiers)...)
	parts = append(parts, prefixedValues("generated_doc", claim.GeneratedDoc)...)
	return strings.Join(parts, "\x00")
}

func verifyChangedClaimImpact(claim claimBinding, changed map[string]bool) (impactClaim, error) {
	record := impactClaim{ID: claim.ID}
	for _, path := range claimRequiredImpactFiles(claim) {
		if changed[path] {
			record.ChangedBoundFiles = append(record.ChangedBoundFiles, path)
		}
	}
	if len(record.ChangedBoundFiles) == 0 {
		return record, fmt.Errorf("claim %s changed without a bound code or test file change", claim.ID)
	}
	record.ChangedReasoningEvidence = changedValues(claimReasoningEvidenceFiles(), changed)
	if len(record.ChangedReasoningEvidence) == 0 {
		return record, fmt.Errorf("claim %s changed without evidence graph reasoning change", claim.ID)
	}
	return record, nil
}

func verifyBoundSurfaceImpact(claim claimBinding, changed map[string]bool) (impactBoundSurface, error) {
	record := impactBoundSurface{ID: claim.ID}
	record.ChangedBoundFiles = changedValues(claim.Files, changed)
	if len(record.ChangedBoundFiles) == 0 {
		return record, nil
	}
	record.ChangedEvidence = changedValues(claimEvidenceFiles(claim), changed)
	if len(record.ChangedEvidence) == 0 {
		return record, fmt.Errorf("claim %s bound files changed without claim evidence change", claim.ID)
	}
	record.ChangedReasoningEvidence = changedValues(claimReasoningEvidenceFiles(), changed)
	if len(record.ChangedReasoningEvidence) == 0 {
		return record, fmt.Errorf("claim %s bound files changed without evidence graph reasoning change", claim.ID)
	}
	return record, nil
}

func claimEvidenceFiles(claim claimBinding) []string {
	values := []string{defaultManifest}
	values = append(values, claim.GeneratedDoc...)
	return values
}

func claimReasoningEvidenceFiles() []string {
	return []string{evidenceGraphManifest}
}
