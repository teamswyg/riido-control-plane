package main

import "fmt"

func verifySourceCoverage(root string, registry loopRegistry, sources []promotionSource) error {
	byLoop := map[string]promotionSource{}
	for _, source := range sources {
		if previous, ok := byLoop[source.HarnessLoop]; ok {
			return fmt.Errorf(
				"harness loop %s has duplicate promotion sources %s and %s",
				source.HarnessLoop,
				previous.ID,
				source.ID,
			)
		}
		byLoop[source.HarnessLoop] = source
	}
	for _, source := range sources {
		loop, ok := findRegisteredLoop(registry, source.HarnessLoop)
		if ok && loop.Kind == "harness" {
			uses, err := loopUsesHarnessPromotion(root, loop)
			if err != nil {
				return err
			}
			if !uses {
				return fmt.Errorf("harness loop %s has sidecar source but workflow does not run harnesspromotion", loop.ID)
			}
		}
	}
	for _, loop := range registry.Loops {
		if loop.Kind != "harness" {
			continue
		}
		uses, err := loopUsesHarnessPromotion(root, loop)
		if err != nil {
			return err
		}
		if !uses {
			continue
		}
		source, ok := byLoop[loop.ID]
		if !ok {
			return fmt.Errorf("harness loop %s has no promotion source", loop.ID)
		}
		if source.SourceWorkflow != loop.RefreshWorkflow {
			return fmt.Errorf("harness loop %s promotion source workflow drift", loop.ID)
		}
	}
	return nil
}
