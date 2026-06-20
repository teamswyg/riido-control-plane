package main

import (
	"encoding/json"
	"os"
	"strings"
)

func contractOwnerDeclaresArtifact(root string, artifact contractArtifact) bool {
	data, err := os.ReadFile(resolvePath(root, artifact.OwnerManifest))
	if err != nil {
		return false
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return false
	}
	value, ok := contractOwnerValue(object, artifact.OwnerKey)
	return ok && value == artifact.Path
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

func contractOwnerHasStrictEvidence(root, ownerManifest string) bool {
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
