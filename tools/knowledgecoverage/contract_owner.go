package main

import "strings"

func ownerManifestDeclaresPath(root, ownerManifest, ownerKey, path string) bool {
	object, ok := readJSONObject(root, ownerManifest)
	if !ok {
		return false
	}
	return ownerValueContainsPath(object, ownerKey, path)
}

func ownerValueContainsPath(object map[string]any, key, path string) bool {
	value, ok := contractOwnerValue(object, key)
	if !ok {
		return false
	}
	return ownerValueMatchesPath(value, path)
}

func contractOwnerValue(object map[string]any, key string) (any, bool) {
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
	return current, true
}

func ownerValueMatchesPath(value any, path string) bool {
	if text, ok := value.(string); ok {
		return text == path
	}
	list, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range list {
		if text, ok := item.(string); ok && text == path {
			return true
		}
	}
	return false
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
