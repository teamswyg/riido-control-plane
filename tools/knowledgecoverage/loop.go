package main

import (
	"encoding/json"
	"os"
	"strings"
)

func manifestHasLoop(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var raw struct {
		Loop evidenceLoop `json:"loop"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}
	return completeLoop(raw.Loop)
}

func completeLoop(loop evidenceLoop) bool {
	values := []string{
		loop.Observation,
		loop.Hypothesis,
		loop.Execute,
		loop.Evaluate,
		loop.Retrospective,
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
