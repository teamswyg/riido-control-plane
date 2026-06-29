package main

import "fmt"

func captureCurrentChainImpact(
	evidence *impactEvidence,
	baseByID map[string]chain,
	current []chain,
	changed map[string]bool,
) error {
	for _, item := range current {
		base, ok := baseByID[item.ID]
		if !ok {
			record, err := verifyChainSurfaceChanged(item, changed, "added")
			if err != nil {
				return err
			}
			evidence.AddedChains = append(evidence.AddedChains, record)
			continue
		}
		if chainSignature(base) == chainSignature(item) {
			continue
		}
		record, err := verifyChainSurfaceChanged(item, changed, "changed")
		if err != nil {
			return err
		}
		evidence.ChangedChains = append(evidence.ChangedChains, record)
	}
	return nil
}

func captureRemovedChainImpact(
	evidence *impactEvidence,
	currentByID map[string]chain,
	base []chain,
	changed map[string]bool,
) error {
	for _, item := range base {
		if _, ok := currentByID[item.ID]; ok {
			continue
		}
		record, err := verifyChainSurfaceChanged(item, changed, "removed")
		if err != nil {
			return err
		}
		evidence.RemovedChains = append(evidence.RemovedChains, record)
	}
	return nil
}

func verifyChainSurfaceChanged(item chain, changed map[string]bool, action string) (impactChain, error) {
	record := impactChain{
		ID:                    item.ID,
		ChangedExecutableRefs: changedChainExecutableRefs(item, changed),
		Claims:                sortedValues(item.Claims),
		VerifierRefs:          refPaths(item.Verifiers),
		EvidenceRefs:          refPaths(item.Evidence),
		NextLoop:              item.NextLoop,
	}
	if len(record.ChangedExecutableRefs) == 0 {
		return record, fmt.Errorf("chain %s %s without executable ref change", item.ID, action)
	}
	return record, nil
}
