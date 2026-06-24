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
	refs := 0
	for _, source := range m.Sources {
		if err := verifySource(root, source); err != nil {
			return verifyResult{}, err
		}
		refs += len(source.RequiredNextArtifacts)
	}
	return verifyResult{SourceCount: len(m.Sources), RequiredRefs: refs}, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected candidate intake identity")
	}
	if m.EvidenceTool != "tools/closedloopcandidateintake" {
		return fmt.Errorf("candidate intake evidence_tool must be tools/closedloopcandidateintake")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("candidate intake must bind generated doc, workflow, and artifact")
	}
	if len(m.Sources) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("candidate intake must declare sources and assertions")
	}
	return nil
}
