package main

import "fmt"

func verifyAll(root string, m manifest, hashes map[string]string) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	loopIDs, result, err := verifyLoops(root, m)
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifyClaims(root, m, loopIDs, hashes); err != nil {
		return verifyResult{}, err
	}
	if err := verifyEvidenceGraph(m, loopIDs); err != nil {
		return verifyResult{}, err
	}
	if err := verifyEvidenceLoop(m.Loop); err != nil {
		return verifyResult{}, err
	}
	if err := verifyRegistryWorkflowCoversClaims(root, m); err != nil {
		return verifyResult{}, err
	}
	result.Claims = len(m.Claims)
	result.GraphEdges = len(m.EvidenceGraph)
	result.Hashes = hashes
	result.ClaimSurfaces = claimSurfaces(m.Claims)
	return result, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected loop registry identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("loop registry must bind generated doc, workflow, and artifact")
	}
	if m.EvidenceTool != "tools/loopregistry" {
		return fmt.Errorf("loop registry evidence_tool must be tools/loopregistry")
	}
	if len(m.Assertions) == 0 || len(m.Loops) == 0 || len(m.Claims) == 0 {
		return fmt.Errorf("loop registry must declare assertions, loops, and claims")
	}
	return nil
}
