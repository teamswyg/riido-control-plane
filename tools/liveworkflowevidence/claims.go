package main

func claimIDs(claims []claimSpec) []string {
	ids := make([]string, 0, len(claims))
	for _, claim := range claims {
		ids = append(ids, claim.ID)
	}
	return ids
}

func newLiveClaims(spec workflowSpec, liveStatus string) []liveClaim {
	if !summaryAllowsField(spec, "evidence_claims") {
		return nil
	}
	claims := make([]liveClaim, 0, len(spec.EvidenceClaims))
	for _, claim := range spec.EvidenceClaims {
		claims = append(claims, liveClaim{
			ID:      claim.ID,
			Summary: claim.Summary,
			Status:  claimStatus(liveStatus),
		})
	}
	return claims
}

func claimStatus(liveStatus string) string {
	if liveStatus == "success" {
		return "verified"
	}
	return "not_verified"
}

func summaryAllowsField(spec workflowSpec, field string) bool {
	for _, allowed := range spec.AllowedFields {
		if allowed == field {
			return true
		}
	}
	return false
}
