package main

func verifyCurrentClaimImpacts(
	evidence *impactEvidence,
	baseByID map[string]claimBinding,
	currentClaims []claimBinding,
	changed map[string]bool,
) error {
	var firstErr error
	for _, claim := range currentClaims {
		base, ok := baseByID[claim.ID]
		if !ok {
			firstErr = rememberImpactError(firstErr,
				recordAddedClaimImpact(evidence, claim, changed))
			continue
		}
		if claimSignature(base) != claimSignature(claim) {
			firstErr = rememberImpactError(firstErr,
				recordChangedClaimImpact(evidence, claim, changed))
		}
		firstErr = rememberImpactError(firstErr,
			recordBoundSurfaceImpact(evidence, claim, changed))
	}
	return firstErr
}

func recordChangedClaimImpact(
	evidence *impactEvidence,
	claim claimBinding,
	changed map[string]bool,
) error {
	record, err := verifyChangedClaimImpact(claim, changed)
	if err != nil {
		evidence.Violations = append(evidence.Violations,
			impactViolationForClaim("changed_claim", claim, record, err))
		return err
	}
	evidence.Claims = append(evidence.Claims, record)
	return nil
}
