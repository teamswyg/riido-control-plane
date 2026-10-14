package main

import "fmt"

type verifyResult struct {
	PublicGates      int `json:"public_gates"`
	PullRequestGates int `json:"pull_request_gates"`
	OperatorGates    int `json:"operator_gates"`
	PrivateGates     int `json:"private_gates"`
	WorkflowRefs     int `json:"workflow_refs"`
	CommandRefs      int `json:"command_refs"`
}

func verifyAll(repoRoot string, m manifest) (verifyResult, error) {
	if err := verifyManifestShape(m); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{PublicGates: len(m.PublicGates), PrivateGates: len(m.PrivateGates)}
	if err := verifyPublicGates(repoRoot, m, &result); err != nil {
		return verifyResult{}, err
	}
	if err := verifyPrivateGates(m.PrivateGates); err != nil {
		return verifyResult{}, err
	}
	return result, nil
}

func verifyManifestShape(m manifest) error {
	if m.SchemaVersion == "" || m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return fmt.Errorf("schema_version, id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.Evidence == "" || len(m.PublicGates) == 0 || len(m.PrivateGates) == 0 {
		return fmt.Errorf("workflow, evidence_artifact, public_gates, and private_gates are required")
	}
	return verifyLoop(m.Loop)
}
