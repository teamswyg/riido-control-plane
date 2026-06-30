package main

func verifyRemovedClaimImpacts(
	evidence *impactEvidence,
	currentByID map[string]claimBinding,
	baseClaims []claimBinding,
	changed map[string]bool,
) error {
	var firstErr error
	for _, claim := range baseClaims {
		if _, ok := currentByID[claim.ID]; ok {
			continue
		}
		firstErr = rememberImpactError(firstErr,
			recordRemovedClaimImpact(evidence, claim, changed))
	}
	return firstErr
}

func recordRemovedClaimImpact(
	evidence *impactEvidence,
	claim claimBinding,
	changed map[string]bool,
) error {
	record, err := verifyRemovedClaimImpact(claim, changed)
	if err != nil {
		evidence.Violations = append(evidence.Violations,
			impactViolationForClaim("removed_claim", claim, record, err))
		return err
	}
	evidence.RemovedClaims = append(evidence.RemovedClaims, record)
	return nil
}
