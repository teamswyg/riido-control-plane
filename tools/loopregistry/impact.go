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
	currentByID := claimsByID(currentClaims)
	evidence := &impactEvidence{
		Enabled:          true,
		BaseRef:          baseRef,
		ChangedFileCount: len(changed),
		ChangedFiles:     changedFileList(changed),
	}
	var firstErr error
	firstErr = rememberImpactError(firstErr,
		verifyCurrentClaimImpacts(evidence, baseByID, currentClaims, changed))
	firstErr = rememberImpactError(firstErr,
		verifyRemovedClaimImpacts(evidence, currentByID, baseClaims, changed))
	evidence.AddedClaimCount = len(evidence.AddedClaims)
	evidence.ChangedClaimCount = len(evidence.Claims)
	evidence.RemovedClaimCount = len(evidence.RemovedClaims)
	evidence.BoundSurfaceChangeCount = len(evidence.BoundSurfaces)
	return evidence, firstErr
}

func rememberImpactError(firstErr, err error) error {
	if firstErr != nil {
		return firstErr
	}
	return err
}
