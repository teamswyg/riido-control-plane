package main

import "strings"

func changedClaimImpactMessage(impact *impactEvidence) string {
	parts := make([]string, 0, 3)
	if claims := impactClaimSummaries(impact.AddedClaims); len(claims) > 0 {
		parts = append(parts, "added claims: "+strings.Join(claims, "; "))
	}
	if claims := impactClaimSummaries(impact.Claims); len(claims) > 0 {
		parts = append(parts, "changed claims: "+strings.Join(claims, "; "))
	}
	if claims := impactClaimSummaries(impact.RemovedClaims); len(claims) > 0 {
		parts = append(parts, "removed claims: "+strings.Join(claims, "; "))
	}
	return strings.Join(parts, "; ")
}

func impactClaimSummaries(claims []impactClaim) []string {
	out := make([]string, 0, len(claims))
	for _, claim := range claims {
		out = append(out, impactClaimSummary(claim))
	}
	return out
}

func impactClaimSummary(claim impactClaim) string {
	parts := []string{claim.ID}
	if len(claim.ChangedBoundFiles) > 0 {
		parts = append(parts, "files: "+strings.Join(claim.ChangedBoundFiles, ", "))
	}
	if len(claim.ChangedEvidence) > 0 {
		parts = append(parts, "evidence: "+strings.Join(claim.ChangedEvidence, ", "))
	}
	if len(claim.ChangedReasoningEvidence) > 0 {
		parts = append(parts, "reasoning: "+strings.Join(claim.ChangedReasoningEvidence, ", "))
	}
	return strings.Join(parts, " ")
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
