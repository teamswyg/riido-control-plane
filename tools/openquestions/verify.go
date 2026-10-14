package main

import "fmt"

func verifyAll(m manifest) (verifyResult, error) {
	if err := verifyShape(m); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{QuestionCount: len(m.Questions), StatusCounts: map[string]int{}}
	seen := map[string]bool{}
	for _, item := range m.Questions {
		if err := verifyQuestion(item, seen, &result); err != nil {
			return verifyResult{}, err
		}
	}
	return result, nil
}

func verifyShape(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version must be %s", manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		return fmt.Errorf("id, title, generated_doc, and workflow are required")
	}
	if m.Evidence == "" || len(m.Questions) == 0 {
		return fmt.Errorf("evidence_artifact and questions are required")
	}
	return verifyLoop(m.Loop)
}
