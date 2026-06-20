package main

import (
	"encoding/json"
	"os"
)

func importedOwnerPointerMatches(root string, imported importedManifest) bool {
	object, ok := readJSONObject(root, imported.OwnerManifest)
	if !ok {
		return false
	}
	pointer, ok := ownerObjectValue(object, imported.OwnerKey)
	return ok &&
		ownerString(pointer, "repo") == imported.UpstreamRepo &&
		ownerString(pointer, "path") == imported.UpstreamPath &&
		ownerString(pointer, "schema_version") == imported.SchemaVersion &&
		ownerString(pointer, "id") == imported.ID
}

func importedLocalMirrorMatches(root string, imported importedManifest) bool {
	object, ok := readJSONObject(root, imported.Path)
	if !ok {
		return false
	}
	return ownerString(object, "schema_version") == imported.SchemaVersion &&
		ownerString(object, "id") == imported.ID
}

func readJSONObject(root, path string) (map[string]any, bool) {
	data, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return nil, false
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, false
	}
	return object, true
}
