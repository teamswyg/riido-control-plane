package main

import "fmt"

func maybeVerifyImpact(root, manifestPath, baseRef string, current manifest) (*impactEvidence, error) {
	if baseRef == "" {
		return nil, nil
	}
	base, err := gitManifest(root, baseRef, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load base evidence graph from %s: %w", baseRef, err)
	}
	changed, err := gitChangedFiles(root, baseRef)
	if err != nil {
		return nil, fmt.Errorf("load changed files from %s: %w", baseRef, err)
	}
	return verifyChainImpact(baseRef, base.Chains, current.Chains, changed)
}

func verifyChainImpact(
	baseRef string,
	baseChains []chain,
	currentChains []chain,
	changed map[string]bool,
) (*impactEvidence, error) {
	baseByID := chainsByID(baseChains)
	currentByID := chainsByID(currentChains)
	evidence := &impactEvidence{Enabled: true, BaseRef: baseRef, ChangedFileCount: len(changed)}
	if err := captureCurrentChainImpact(evidence, baseByID, currentChains, changed); err != nil {
		return nil, err
	}
	if err := captureRemovedChainImpact(evidence, currentByID, baseChains, changed); err != nil {
		return nil, err
	}
	evidence.AddedChainCount = len(evidence.AddedChains)
	evidence.ChangedChainCount = len(evidence.ChangedChains)
	evidence.RemovedChainCount = len(evidence.RemovedChains)
	return evidence, nil
}
