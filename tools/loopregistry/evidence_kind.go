package main

import "fmt"

func verifyEvidenceKinds(m manifest) error {
	kinds, err := evidenceKindVocabulary(m.EvidenceKinds)
	if err != nil {
		return err
	}
	for _, loop := range m.Loops {
		for _, source := range loop.Evidence {
			if !kinds[source.Kind] {
				return fmt.Errorf("loop %s uses unknown evidence kind %s", loop.ID, source.Kind)
			}
		}
	}
	return nil
}

func evidenceKindVocabulary(items []evidenceKind) (map[string]bool, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("evidence_kinds vocabulary is required")
	}
	out := map[string]bool{}
	for _, item := range items {
		if item.Kind == "" || item.Description == "" {
			return nil, fmt.Errorf("evidence kind must bind kind and description")
		}
		if out[item.Kind] {
			return nil, fmt.Errorf("duplicate evidence kind %s", item.Kind)
		}
		out[item.Kind] = true
	}
	return out, nil
}
