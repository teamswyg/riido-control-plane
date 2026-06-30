package main

import "fmt"

func verifyArchitectureComponentFileSet(c architectureComponent) error {
	seen := map[string]bool{}
	for _, file := range c.Files {
		if seen[file] {
			return fmt.Errorf("architecture component %s duplicates file %s", c.ID, file)
		}
		seen[file] = true
	}
	return nil
}
