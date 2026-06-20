package main

import "fmt"

func validateImportedManifests(root string, m manifest) []string {
	var problems []string
	for _, imported := range m.ImportedManifests {
		problems = append(problems, validateImportedManifest(root, imported)...)
	}
	return problems
}

func validateImportedManifest(root string, imported importedManifest) []string {
	if imported.Path == "" || imported.OwnerManifest == "" || imported.OwnerKey == "" {
		return []string{"imported manifest path, owner_manifest, and owner_key are required"}
	}
	if !importedOwnerPointerMatches(root, imported) {
		return []string{fmt.Sprintf("%s owner pointer %q in %q must match imported manifest identity",
			imported.Path, imported.OwnerKey, imported.OwnerManifest)}
	}
	if !importedLocalMirrorMatches(root, imported) {
		return []string{fmt.Sprintf("%s local mirror must match imported manifest identity", imported.Path)}
	}
	if !contractOwnerHasStrictEvidence(root, imported.OwnerManifest) {
		return []string{fmt.Sprintf("%s owner manifest %q must have strict generated evidence",
			imported.Path, imported.OwnerManifest)}
	}
	return nil
}

func importedManifestBindingCount(root string, m manifest) int {
	count := 0
	for _, imported := range m.ImportedManifests {
		if len(validateImportedManifest(root, imported)) == 0 {
			count++
		}
	}
	return count
}

func importedManifestMissingBinding(root string, m manifest) []string {
	var paths []string
	for _, imported := range m.ImportedManifests {
		if len(validateImportedManifest(root, imported)) > 0 {
			paths = append(paths, imported.Path)
		}
	}
	return emptyIfNil(paths)
}
