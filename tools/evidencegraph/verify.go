package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyManifest(m); err != nil {
		return verifyResult{}, err
	}
	nextLoops, err := loadLoopRegistryIDs(root, m.LoopRegistry)
	if err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{Chains: len(m.Chains)}
	seen := map[string]bool{}
	for _, c := range m.Chains {
		if err := verifyChain(root, c, seen, &result, nextLoops); err != nil {
			return verifyResult{}, err
		}
	}
	return result, nil
}

func verifyManifest(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version must be %s", manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return fmt.Errorf("id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.Evidence == "" || m.EvidenceTool != "tools/evidencegraph" {
		return fmt.Errorf("workflow, evidence_artifact, and evidence_tool are required")
	}
	if m.LoopRegistry == "" {
		return fmt.Errorf("loop_registry_manifest is required")
	}
	if len(m.Assertions) == 0 || len(m.Chains) == 0 {
		return fmt.Errorf("assertions and chains are required")
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
