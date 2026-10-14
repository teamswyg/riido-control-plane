package main

import "fmt"

type verifyResult struct {
	Sections        int
	Slices          int
	ValidationGates int
	RiidoReferences int
}

func verify(m manifest) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return verifyResult{}, err
	}
	if err := verifyRequiredSections(m.Sections); err != nil {
		return verifyResult{}, err
	}
	if len(m.Assertions) == 0 || len(m.Intro) == 0 {
		return verifyResult{}, fmt.Errorf("intro and assertions are required")
	}
	result, err := verifySections(m.Sections)
	if err != nil {
		return verifyResult{}, err
	}
	if result.Slices < 80 || result.ValidationGates < 5 || result.RiidoReferences < 50 {
		return verifyResult{}, fmt.Errorf("migration ledger coverage too small: %+v", result)
	}
	return result, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != expectedID || m.RiidoTask != expectedTask {
		return fmt.Errorf("unexpected manifest identity")
	}
	if m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("title, generated_doc, workflow, and evidence_artifact are required")
	}
	return nil
}
