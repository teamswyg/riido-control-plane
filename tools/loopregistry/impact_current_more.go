package main

func recordAddedClaimImpact(
	evidence *impactEvidence,
	claim claimBinding,
	changed map[string]bool,
) error {
	record, err := verifyAddedClaimImpact(claim, changed)
	if err != nil {
		evidence.Violations = append(evidence.Violations,
			impactViolationForClaim("added_claim", claim, record, err))
		return err
	}
	evidence.AddedClaims = append(evidence.AddedClaims, record)
	return nil
}

func recordBoundSurfaceImpact(
	evidence *impactEvidence,
	claim claimBinding,
	changed map[string]bool,
) error {
	record, err := verifyBoundSurfaceImpact(claim, changed)
	if err != nil {
		evidence.Violations = append(evidence.Violations,
			impactViolationForBoundSurface(claim, record, err))
		return err
	}
	if len(record.ChangedBoundFiles) > 0 {
		evidence.BoundSurfaces = append(evidence.BoundSurfaces, record)
	}
	return nil
}
