package main

import (
	"encoding/json"
	"os"
)

type standaloneLoopManifest struct {
	Loop evidenceLoop `json:"loop"`
}

func standaloneManifestHasLoop(root, path string) bool {
	data, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return false
	}
	var manifest standaloneLoopManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return false
	}
	return completeLoop(manifest.Loop)
}
