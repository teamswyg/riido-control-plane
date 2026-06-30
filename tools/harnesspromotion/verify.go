package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return verifyResult{}, err
	}
	registry, err := loadLoopRegistry(root, m.LoopRegistryManifest)
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifySourceCoverage(root, registry, m.Sources); err != nil {
		return verifyResult{}, err
	}
	coverage, err := sourceCoverageSummary(root, registry)
	if err != nil {
		return verifyResult{}, err
	}
	claimCount := 0
	for _, source := range m.Sources {
		if err := verifySource(root, source); err != nil {
			return verifyResult{}, err
		}
		if err := verifySourceLoopBinding(registry, source); err != nil {
			return verifyResult{}, err
		}
		if err := verifyRequiredNextArtifacts(source); err != nil {
			return verifyResult{}, err
		}
		claimCount += len(source.RequiredNextArtifacts)
	}
	return verifyResult{
		SourceCount: len(m.Sources), ClaimCount: claimCount,
		SidecarSourceCount:              coverage.SidecarSourceCount,
		LoopOwnedCandidateProducerCount: coverage.LoopOwnedCandidateProducerCount,
	}, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected harness promotion identity")
	}
	if m.EvidenceTool != "tools/harnesspromotion" {
		return fmt.Errorf("harness promotion evidence_tool must be tools/harnesspromotion")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" || m.LoopRegistryManifest == "" {
		return fmt.Errorf("harness promotion must bind generated doc, workflow, artifact, and loop registry")
	}
	if len(m.Sources) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("harness promotion must declare sources and assertions")
	}
	return nil
}
