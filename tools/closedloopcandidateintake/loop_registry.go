package main

import "fmt"

type loopRegistry struct {
	Loops []registeredLoop `json:"loops"`
}

type registeredLoop struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func verifyLoopRegistry(root string, source intakeSource) error {
	var registry loopRegistry
	if err := readJSON(repoPath(root, source.LoopRegistryManifest), &registry); err != nil {
		return err
	}
	for _, loop := range registry.Loops {
		if loop.ID == source.PromotionTarget {
			if loop.Kind != "closed_loop" {
				return fmt.Errorf("promotion target %s is not closed_loop", source.PromotionTarget)
			}
			return nil
		}
	}
	return fmt.Errorf("promotion target %s missing in loop registry", source.PromotionTarget)
}
