package main

import "strings"

func changedClaimImpactMessage(impact *impactEvidence) string {
	parts := make([]string, 0, 3)
	if ids := impactClaimIDs(impact.AddedClaims); len(ids) > 0 {
		parts = append(parts, "added claims: "+strings.Join(ids, ", "))
	}
	if ids := impactClaimIDs(impact.Claims); len(ids) > 0 {
		parts = append(parts, "changed claims: "+strings.Join(ids, ", "))
	}
	if ids := impactClaimIDs(impact.RemovedClaims); len(ids) > 0 {
		parts = append(parts, "removed claims: "+strings.Join(ids, ", "))
	}
	return strings.Join(parts, "; ")
}

func impactClaimIDs(claims []impactClaim) []string {
	ids := make([]string, 0, len(claims))
	for _, claim := range claims {
		ids = append(ids, claim.ID)
	}
	return ids
}

func boundSurfaceImpactMessage(impact *impactEvidence) string {
	if len(impact.BoundSurfaces) == 0 {
		return ""
	}
	parts := make([]string, 0, len(impact.BoundSurfaces))
	for _, surface := range impact.BoundSurfaces {
		parts = append(parts, boundSurfaceSummary(surface))
	}
	return "bound surfaces: " + strings.Join(parts, "; ")
}

func boundSurfaceSummary(surface impactBoundSurface) string {
	parts := []string{surface.ID}
	if len(surface.ChangedBoundFiles) > 0 {
		parts = append(parts, "files: "+strings.Join(surface.ChangedBoundFiles, ", "))
	}
	if len(surface.ChangedEvidence) > 0 {
		parts = append(parts, "evidence: "+strings.Join(surface.ChangedEvidence, ", "))
	}
	if len(surface.ChangedReasoningEvidence) > 0 {
		parts = append(parts, "reasoning: "+strings.Join(surface.ChangedReasoningEvidence, ", "))
	}
	return strings.Join(parts, " ")
}
