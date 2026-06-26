package main

import "fmt"

func verifyQuestion(item question, seen map[string]bool, result *verifyResult) error {
	if item.ID == "" || item.Area == "" || item.Owner == "" || item.Question == "" || item.Stance == "" {
		return fmt.Errorf("question id, area, owner, question, and stance are required")
	}
	if seen[item.ID] {
		return fmt.Errorf("duplicate question id %q", item.ID)
	}
	seen[item.ID] = true
	result.StatusCounts[item.Status]++
	switch item.Status {
	case "open":
		result.OpenCount++
		if item.NextArtifact == "" || item.NextArtifact == "none" {
			return fmt.Errorf("open question %s requires a next artifact", item.ID)
		}
		if err := verifyNextCommand(item); err != nil {
			return err
		}
	case "resolved-no-diff", "resolved":
		result.ResolvedCount++
		if item.Resolution == "" {
			return fmt.Errorf("resolved question %s requires a resolution", item.ID)
		}
	default:
		return fmt.Errorf("unknown question status %q", item.Status)
	}
	return nil
}
