package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func verifyLinkedCDManifest(repoRoot string, m manifest) error {
	body, err := os.ReadFile(repoPath(repoRoot, m.LinkedCD))
	if err != nil {
		return fmt.Errorf("read linked runtime CD manifest: %w", err)
	}
	var linked struct {
		SchemaVersion string `json:"schema_version"`
		ID            string `json:"id"`
		Current       struct {
			Workflow string `json:"workflow"`
		} `json:"current_strategy"`
	}
	if err := json.Unmarshal(body, &linked); err != nil {
		return fmt.Errorf("decode linked runtime CD manifest: %w", err)
	}
	if linked.SchemaVersion != "riido-control-plane-runtime-cd-ownership.v1" {
		return fmt.Errorf("linked runtime CD manifest has unexpected schema")
	}
	if linked.ID != "runtime-cd-ownership" {
		return fmt.Errorf("linked runtime CD manifest has unexpected id %q", linked.ID)
	}
	if linked.Current.Workflow != ".github/workflows/deploy-ai-agent-testnet.yml" {
		return fmt.Errorf("linked runtime CD workflow changed to %q", linked.Current.Workflow)
	}
	return nil
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
