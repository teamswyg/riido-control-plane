package main

import "fmt"

func verify(root string, m manifest, checkDoc bool) error {
	if err := verifyHeader(m); err != nil {
		return err
	}
	if err := verifyContractMirror(root, m); err != nil {
		return err
	}
	if err := verifyStaticLists(m); err != nil {
		return err
	}
	if err := verifySources(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyHeader(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != expectedID || m.RiidoTask != expectedTask {
		return fmt.Errorf("unexpected manifest identity")
	}
	required := []string{m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact}
	for _, value := range required {
		if value == "" {
			return fmt.Errorf("title, generated_doc, workflow, and evidence_artifact are required")
		}
	}
	return nil
}
