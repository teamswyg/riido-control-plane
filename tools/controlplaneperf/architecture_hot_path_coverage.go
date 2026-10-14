package main

import "fmt"

func verifyArchitectureHotPathFiles(m manifest) error {
	covered := architectureComponentFiles(m.ArchitectureComponents)
	for _, path := range m.HotPaths {
		for _, file := range path.Files {
			if !covered[file] {
				return fmt.Errorf("hot path %s file %s lacks architecture component", path.ID, file)
			}
		}
	}
	return nil
}

func architectureComponentFiles(components []architectureComponent) map[string]bool {
	out := map[string]bool{}
	for _, component := range components {
		for _, file := range component.Files {
			out[file] = true
		}
	}
	return out
}
