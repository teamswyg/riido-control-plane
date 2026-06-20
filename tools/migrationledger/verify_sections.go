package main

import "fmt"

func verifyRequiredSections(sections []section) error {
	required := []string{
		"Goal", "Source In The Private Repository", "Target Boundary",
		"Migration Order", "Current Migration Slices", "Validation Gates",
		"Infra Boundary", "Contract Boundary",
	}
	for _, title := range required {
		if !hasSection(sections, title) {
			return fmt.Errorf("missing required section %q", title)
		}
	}
	return nil
}

func hasSection(sections []section, title string) bool {
	for _, item := range sections {
		if item.Title == title {
			return true
		}
	}
	return false
}

func verifySections(sections []section) (verifyResult, error) {
	result := verifyResult{Sections: len(sections)}
	for _, item := range sections {
		if err := verifySectionShape(item); err != nil {
			return verifyResult{}, err
		}
		result.add(item)
	}
	return result, nil
}

func verifySectionShape(item section) error {
	if item.Level < 2 || item.Level > 3 || item.Title == "" || item.Kind == "" {
		return fmt.Errorf("invalid section shape %+v", item)
	}
	if item.Level == 3 && len(item.Body) == 0 {
		return fmt.Errorf("slice section %q must keep body evidence", item.Title)
	}
	return nil
}
