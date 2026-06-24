package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyManifest(m); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{Hooks: len(m.Hooks), Scripts: len(m.Scripts)}
	if err := verifyPreCommitConfig(root, m, &result); err != nil {
		return verifyResult{}, err
	}
	if err := verifyScripts(root, m, &result); err != nil {
		return verifyResult{}, err
	}
	return result, nil
}

func verifyManifest(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version must be %s", manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.Evidence == "" {
		return fmt.Errorf("id, title, generated_doc, workflow, and evidence_artifact are required")
	}
	if m.PreCommitConfig == "" || len(m.Hooks) == 0 || len(m.Scripts) == 0 {
		return fmt.Errorf("pre_commit_config, hooks, and scripts are required")
	}
	return verifyLoop(m.Loop)
}

func verifyLoop(loop loopRecord) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" {
		return fmt.Errorf("loop observation, hypothesis, and execute are required")
	}
	if loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("loop evaluate and retrospective are required")
	}
	return nil
}
