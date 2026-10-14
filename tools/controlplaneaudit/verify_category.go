package main

import "fmt"

func verifyRequiredCategories(required []string, surfaces []surface) error {
	present := map[string]bool{}
	for _, surface := range surfaces {
		present[surface.Category] = true
	}
	for _, category := range required {
		if !present[category] {
			return fmt.Errorf("audit missing required high-traffic category %s", category)
		}
	}
	return nil
}
