package main

import (
	"encoding/json"
	"os"
)

func loadGeneratedManifestMeta(root, path string) (generatedManifestMeta, bool) {
	data, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return generatedManifestMeta{}, false
	}
	var meta generatedManifestMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return generatedManifestMeta{}, false
	}
	return meta, true
}

func generatedToolForDoc(root, path string) string {
	data, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return ""
	}
	return generatedToolFromMarker(string(data))
}
