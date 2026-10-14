package main

import (
	"encoding/json"
	"os"
	"strings"
)

type directManifestMeta struct {
	Loop         evidenceLoop `json:"loop"`
	EvidenceTool string       `json:"evidence_tool"`
}

func directManifestMetadata(path string) directManifestMeta {
	data, err := os.ReadFile(path)
	if err != nil {
		return directManifestMeta{}
	}
	var raw directManifestMeta
	if err := json.Unmarshal(data, &raw); err != nil {
		return directManifestMeta{}
	}
	return raw
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
