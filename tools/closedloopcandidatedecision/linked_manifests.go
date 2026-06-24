package main

import "fmt"

type intakeManifest struct {
	Sources []intakeSource `json:"sources"`
}

type intakeSource struct {
	PromotionTarget string `json:"promotion_target"`
}

func verifyLinkedManifests(root string, m manifest) error {
	var intake intakeManifest
	if err := readJSON(repoPath(root, m.IntakeManifest), &intake); err != nil {
		return err
	}
	if len(intake.Sources) == 0 {
		return fmt.Errorf("candidate decision intake manifest has no sources")
	}
	var registry loopRegistry
	if err := readJSON(repoPath(root, m.LoopRegistryManifest), &registry); err != nil {
		return err
	}
	for _, decision := range m.Decisions {
		if !registryHasLoop(registry, decision.NextLoop) {
			return fmt.Errorf("candidate %s references unknown next loop %s", decision.CandidateID, decision.NextLoop)
		}
	}
	return nil
}
