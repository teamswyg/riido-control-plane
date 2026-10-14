package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type loopRegistry struct {
	Loops []registeredLoop `json:"loops"`
}

type registeredLoop struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	RefreshWorkflow string   `json:"refresh_workflow"`
	PromotesTo      []string `json:"promotes_to,omitempty"`
}

func loadLoopRegistry(root, path string) (loopRegistry, error) {
	var registry loopRegistry
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return registry, fmt.Errorf("read loop registry %s: %w", path, err)
	}
	return registry, json.Unmarshal(data, &registry)
}

func verifySourceLoopBinding(registry loopRegistry, source promotionSource) error {
	harness, ok := findRegisteredLoop(registry, source.HarnessLoop)
	if !ok || harness.Kind != "harness" {
		return fmt.Errorf("source %s references non-harness loop %s", source.ID, source.HarnessLoop)
	}
	target, ok := findRegisteredLoop(registry, source.PromotionTarget)
	if !ok || target.Kind != "closed_loop" {
		return fmt.Errorf("source %s references non-closed-loop target %s", source.ID, source.PromotionTarget)
	}
	if !containsString(harness.PromotesTo, source.PromotionTarget) {
		return fmt.Errorf("source %s target %s is not declared by harness loop", source.ID, source.PromotionTarget)
	}
	return nil
}

func findRegisteredLoop(registry loopRegistry, id string) (registeredLoop, bool) {
	for _, loop := range registry.Loops {
		if loop.ID == id {
			return loop, true
		}
	}
	return registeredLoop{}, false
}
