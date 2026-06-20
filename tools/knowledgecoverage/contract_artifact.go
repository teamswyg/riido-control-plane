package main

import "fmt"

func validateContractArtifacts(root string, m manifest) []string {
	var problems []string
	for _, artifact := range m.ContractArtifacts {
		problems = append(problems, validateContractArtifact(root, artifact)...)
	}
	return problems
}

func validateContractArtifact(root string, artifact contractArtifact) []string {
	if artifact.Path == "" || artifact.OwnerManifest == "" || artifact.OwnerKey == "" {
		return []string{"contract artifact path, owner_manifest, and owner_key are required"}
	}
	if !contractOwnerDeclaresArtifact(root, artifact) {
		return []string{fmt.Sprintf("%s owner manifest %q key %q must declare artifact path",
			artifact.Path, artifact.OwnerManifest, artifact.OwnerKey)}
	}
	if !contractOwnerHasStrictEvidence(root, artifact.OwnerManifest) {
		return []string{fmt.Sprintf("%s owner manifest %q must have strict generated evidence",
			artifact.Path, artifact.OwnerManifest)}
	}
	return nil
}

func contractArtifactBindingCount(root string, m manifest) int {
	count := 0
	for _, artifact := range m.ContractArtifacts {
		if len(validateContractArtifact(root, artifact)) == 0 {
			count++
		}
	}
	return count
}

func contractArtifactMissingBinding(root string, m manifest) []string {
	var paths []string
	for _, artifact := range m.ContractArtifacts {
		if len(validateContractArtifact(root, artifact)) > 0 {
			paths = append(paths, artifact.Path)
		}
	}
	return emptyIfNil(paths)
}
