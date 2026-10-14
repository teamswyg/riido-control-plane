package main

import "fmt"

func verify(root string, m manifest, checkDoc bool) error {
	if err := verifyHeader(m); err != nil {
		return err
	}
	if err := verifyContexts(root, m); err != nil {
		return err
	}
	if err := verifyLinks(root, m.SSOTLinks); err != nil {
		return err
	}
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyForbiddenImports(root, m.DirectionRules); err != nil {
		return err
	}
	if err := verifyLoop(m.Loop); err != nil {
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
	for _, value := range []string{m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact} {
		if value == "" {
			return fmt.Errorf("title, generated_doc, workflow, and evidence_artifact are required")
		}
	}
	return nil
}
