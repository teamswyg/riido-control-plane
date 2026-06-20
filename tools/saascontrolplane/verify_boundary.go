package main

import "fmt"

func verifyBoundaries(repo string, m manifest) error {
	if len(m.Boundaries) < 12 {
		return fmt.Errorf("boundary coverage is underspecified: %d", len(m.Boundaries))
	}
	seen := map[string]bool{}
	for _, item := range m.Boundaries {
		if item.ID == "" || item.Summary == "" {
			return fmt.Errorf("boundary identity is incomplete: %+v", item)
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate boundary %q", item.ID)
		}
		seen[item.ID] = true
		if err := verifyBoundaryWorkflow(repo, m, item); err != nil {
			return err
		}
		if len(item.SourceChecks) == 0 {
			return fmt.Errorf("boundary %q has no source checks", item.ID)
		}
		for _, check := range item.SourceChecks {
			if err := verifySourceCheck(repo, item.ID, check); err != nil {
				return err
			}
		}
	}
	return nil
}
