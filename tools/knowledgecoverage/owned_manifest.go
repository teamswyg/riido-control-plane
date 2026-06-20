package main

import "fmt"

func validateOwnedManifests(root string, m manifest) []string {
	var problems []string
	for _, owned := range m.OwnedManifests {
		problems = append(problems, validateOwnedManifest(root, owned)...)
	}
	return problems
}

func validateOwnedManifest(root string, owned ownedManifest) []string {
	if owned.Path == "" || owned.OwnerManifest == "" || owned.OwnerKey == "" {
		return []string{"owned manifest path, owner_manifest, and owner_key are required"}
	}
	if !ownerManifestDeclaresPath(root, owned.OwnerManifest, owned.OwnerKey, owned.Path) {
		return []string{fmt.Sprintf("%s owner manifest %q key %q must declare owned path",
			owned.Path, owned.OwnerManifest, owned.OwnerKey)}
	}
	if !ownerManifestHasStrictEvidence(root, owned.OwnerManifest) {
		return []string{fmt.Sprintf("%s owner manifest %q must have strict generated evidence",
			owned.Path, owned.OwnerManifest)}
	}
	return nil
}

func ownedManifestBindingCount(root string, m manifest) int {
	count := 0
	for _, owned := range m.OwnedManifests {
		if len(validateOwnedManifest(root, owned)) == 0 {
			count++
		}
	}
	return count
}

func ownedManifestMissingBinding(root string, m manifest) []string {
	var paths []string
	for _, owned := range m.OwnedManifests {
		if len(validateOwnedManifest(root, owned)) > 0 {
			paths = append(paths, owned.Path)
		}
	}
	return emptyIfNil(paths)
}
