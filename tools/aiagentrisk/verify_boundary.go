package main

import (
	"fmt"
	"slices"
)

func verifyBoundaries(boundaries []remainingBoundary) error {
	seen := map[string]bool{}
	for _, boundary := range boundaries {
		if boundary.ID == "" || boundary.Owner == "" || boundary.Reason == "" {
			return fmt.Errorf("invalid remaining boundary %+v", boundary)
		}
		if !slices.Contains(requiredBoundaries, boundary.ID) {
			return fmt.Errorf("unexpected remaining boundary %q", boundary.ID)
		}
		seen[boundary.ID] = true
	}
	for _, boundary := range requiredBoundaries {
		if !seen[boundary] {
			return fmt.Errorf("manifest missing remaining boundary %q", boundary)
		}
	}
	return nil
}
