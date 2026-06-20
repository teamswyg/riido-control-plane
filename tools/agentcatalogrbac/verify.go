package main

import "fmt"

func verify(root string, m manifest, checkDoc bool) error {
	if err := verifyHeader(m); err != nil {
		return err
	}
	if err := verifyDomain(m); err != nil {
		return err
	}
	if err := verifyProfiles(root, m); err != nil {
		return err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return err
	}
	if err := verifySources(root, m); err != nil {
		return err
	}
	if checkDoc {
		if err := verifyDoc(root, m); err != nil {
			return err
		}
	}
	return nil
}

func verifyHeader(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != expectedID || m.RiidoTask != expectedTask {
		return fmt.Errorf("unexpected manifest identity")
	}
	for _, value := range []string{m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact, m.OwnerPackage} {
		if value == "" {
			return fmt.Errorf("title, generated_doc, workflow, evidence_artifact, and owner_package are required")
		}
	}
	return nil
}
