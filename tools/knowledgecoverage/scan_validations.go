package main

func scanValidationProblems(root string, m manifest, docs []docClass) []string {
	var problems []string
	problems = append(problems, validateGeneratedDocs(root, docs)...)
	problems = append(problems, validateGeneratedArtifactBindings(root, docs)...)
	problems = append(problems, validateManualEntries(root, m, docs)...)
	problems = append(problems, validateDirectLoops(docs)...)
	problems = append(problems, validateDirectEvidence(root, docs)...)
	problems = append(problems, validateStandaloneManifests(root, m)...)
	problems = append(problems, validateSourceManifests(root, m)...)
	problems = append(problems, validateContractArtifacts(root, m)...)
	problems = append(problems, validateImportedManifests(root, m)...)
	problems = append(problems, validateOwnedManifests(root, m)...)
	problems = append(problems, validateManifestInventory(root, m, docs)...)
	return problems
}
