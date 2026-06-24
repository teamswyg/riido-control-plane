package main

import "fmt"

func maybeVerifyImpact(root, manifestPath, baseRef string, current manifest) (*impactEvidence, error) {
	if baseRef == "" {
		return nil, nil
	}
	base, err := gitManifest(root, baseRef, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load base loop registry from %s: %w", baseRef, err)
	}
	changed, err := gitChangedFiles(root, baseRef)
	if err != nil {
		return nil, fmt.Errorf("load changed files from %s: %w", baseRef, err)
	}
	return verifyClaimImpact(baseRef, base.Claims, current.Claims, changed)
}

func verifyClaimImpact(
	baseRef string,
	baseClaims []claimBinding,
	currentClaims []claimBinding,
	changed map[string]bool,
) (*impactEvidence, error) {
	baseByID := claimsByID(baseClaims)
	evidence := &impactEvidence{Enabled: true, BaseRef: baseRef, ChangedFileCount: len(changed)}
	for _, claim := range currentClaims {
		base, ok := baseByID[claim.ID]
		if !ok || claimSignature(base) == claimSignature(claim) {
			continue
		}
		record, err := verifyChangedClaimImpact(claim, changed)
		if err != nil {
			return nil, err
		}
		evidence.Claims = append(evidence.Claims, record)
	}
	evidence.ChangedClaimCount = len(evidence.Claims)
	return evidence, nil
}
