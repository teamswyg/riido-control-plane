package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return verifyResult{}, err
	}
	if err := verifyWorkflow(root, m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLinkedManifests(root, m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyDecisions(m); err != nil {
		return verifyResult{}, err
	}
	return verifyResult{DecisionCount: len(m.Decisions)}, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected candidate decision identity")
	}
	if m.EvidenceTool != "tools/closedloopcandidatedecision" {
		return fmt.Errorf("candidate decision evidence_tool must be tools/closedloopcandidatedecision")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" || m.CommandArtifact == "" {
		return fmt.Errorf("candidate decision must bind generated doc, workflow, evidence artifact, and command artifact")
	}
	return nil
}
