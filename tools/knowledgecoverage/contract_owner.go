package main

import (
	"encoding/json"
	"os"
	"strings"
)

func ownerManifestDeclaresPath(root, ownerManifest, ownerKey, path string) bool {
	data, err := os.ReadFile(resolvePath(root, ownerManifest))
	if err != nil {
		return false
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return false
	}
	value, ok := contractOwnerValue(object, ownerKey)
	return ok && value == path
}

func contractOwnerValue(object map[string]any, key string) (string, bool) {
	var current any = object
	for _, part := range strings.Split(key, ".") {
		next, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = next[part]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}

func ownerManifestHasStrictEvidence(root, ownerManifest string) bool {
	meta, ok := loadGeneratedManifestMeta(root, ownerManifest)
	if !ok {
		return false
	}
	tool := generatedToolForDoc(root, meta.GeneratedDoc)
	if tool == "" {
		return false
	}
	doc := docClass{Path: meta.GeneratedDoc, Kind: "generated", GeneratorTool: tool}
	return len(validateGeneratedManifestBinding(root, doc, meta)) == 0
}
