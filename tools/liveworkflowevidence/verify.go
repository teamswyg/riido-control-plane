package main

import "fmt"

func verifyAll(root string, m manifest) (verifyResult, error) {
	if m.SchemaVersion == "" || m.ID == "" || m.GeneratedDoc == "" {
		return verifyResult{}, fmt.Errorf("manifest identity and generated_doc are required")
	}
	if len(m.Workflows) == 0 {
		return verifyResult{}, fmt.Errorf("at least one live workflow is required")
	}
	seen := map[string]bool{}
	result := verifyResult{}
	for _, spec := range m.Workflows {
		if err := verifyWorkflowSpec(spec, seen); err != nil {
			return verifyResult{}, err
		}
		checks, err := verifyWorkflowFile(root, spec)
		if err != nil {
			return verifyResult{}, err
		}
		result.PhraseChecks += checks
		result.WorkflowCount++
		result.Records = append(result.Records, newRecord(spec))
	}
	return result, nil
}

func verifyWorkflowSpec(spec workflowSpec, seen map[string]bool) error {
	if spec.ID == "" || spec.Path == "" || spec.SummaryArtifact == "" || spec.SummaryPath == "" {
		return fmt.Errorf("workflow identity and summary wiring are required")
	}
	if seen[spec.ID] {
		return fmt.Errorf("duplicate workflow id %q", spec.ID)
	}
	seen[spec.ID] = true
	if len(spec.SensitiveInputs) == 0 || len(spec.AllowedFields) == 0 {
		return fmt.Errorf("workflow %q needs sensitive inputs and allowed fields", spec.ID)
	}
	return nil
}

func findWorkflow(m manifest, id string) (workflowSpec, bool) {
	for _, spec := range m.Workflows {
		if spec.ID == id {
			return spec, true
		}
	}
	return workflowSpec{}, false
}
