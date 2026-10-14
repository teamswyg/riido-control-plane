package main

import "fmt"

func verifyDecisionTemplates(templates []decisionTemplate) error {
	seen := map[string]bool{}
	for _, template := range templates {
		if template.SubjectKind == "" {
			return fmt.Errorf("decision template must name subject_kind")
		}
		if seen[template.SubjectKind] {
			return fmt.Errorf("duplicate decision template for subject kind %s", template.SubjectKind)
		}
		seen[template.SubjectKind] = true
		if err := verifyDecision(decisionFromTemplate("template:"+template.SubjectKind, template)); err != nil {
			return err
		}
	}
	return nil
}
