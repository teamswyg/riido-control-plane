package main

const evidenceSchema = "riido-executable-knowledge-coverage-result.v1"

func buildEvidence(root string, m manifest, docs []docClass, problems []string) evidence {
	counts := countDocs(docs)
	loops := scanManifestLoops(root, m)
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	if problems == nil {
		problems = []string{}
	}
	return evidence{
		SchemaVersion: evidenceSchema, ID: m.ID, Status: status,
		ScannedCount: len(docs), GeneratedCount: counts["generated"],
		GeneratedToolCount:             generatedToolCount(docs),
		GeneratedEvidenceWorkflowCount: generatedEvidenceWorkflowCount(root, docs),
		GeneratedDeclaredWorkflowCount: generatedDeclaredWorkflowEvidenceCount(root, docs),
		GeneratedManifestEvidenceTool:  generatedManifestEvidenceToolCount(root, docs),
		GeneratedArtifactBindingCount:  generatedArtifactBindingCount(root, docs),
		DirectSSOTCount:                counts["direct_ssot"], ManualCount: counts["manual_registered"],
		ManualByGroup: manualCountsByGroup(docs), ManualTopDirs: manualTopDirs(docs, 8),
		ManualSamples: manualSamples(docs, 2), DirectLoopCount: countDirectLoops(docs),
		GeneratedMissingTool:             generatedMissingTool(root, docs),
		GeneratedMissingWorkflow:         generatedMissingWorkflow(root, docs),
		GeneratedMissingEvidenceWorkflow: generatedMissingEvidenceWorkflow(root, docs),
		GeneratedMissingDeclaredWorkflow: generatedMissingDeclaredWorkflowEvidence(root, docs),
		GeneratedEvidenceToolMismatch:    generatedManifestEvidenceToolMismatch(root, docs),
		GeneratedMissingArtifactBinding:  generatedMissingArtifactBinding(root, docs),
		DirectEvidenceWorkflowCount:      directEvidenceWorkflowCount(root, docs),
		StandaloneManifestCount:          len(m.Standalone),
		StandaloneManifestBindingCount:   standaloneManifestBindingCount(root, m),
		StandaloneMissingBinding:         standaloneManifestMissingBinding(root, m),
		SourceManifestCount:              len(m.SourceManifests),
		SourceManifestMetadataCount:      sourceManifestMetadataCount(root, m),
		SourceManifestBindingCount:       sourceManifestBindingCount(root, m),
		SourceMissingMetadata:            sourceManifestMissingMetadata(root, m),
		SourceMissingBinding:             sourceManifestMissingBinding(root, m),
		ContractArtifactCount:            len(m.ContractArtifacts),
		ContractArtifactBindingCount:     contractArtifactBindingCount(root, m),
		ContractMissingBinding:           contractArtifactMissingBinding(root, m),
		ImportedManifestCount:            len(m.ImportedManifests),
		ImportedManifestBindingCount:     importedManifestBindingCount(root, m),
		ImportedMissingBinding:           importedManifestMissingBinding(root, m),
		OwnedManifestCount:               len(m.OwnedManifests),
		OwnedManifestBindingCount:        ownedManifestBindingCount(root, m),
		OwnedMissingBinding:              ownedManifestMissingBinding(root, m),
		ManifestInventoryCount:           manifestInventoryCount(root),
		TrackedManifestCount:             trackedManifestCount(root, m, docs),
		ManifestInventoryByGroup:         manifestInventoryByGroup(root),
		ManifestInventorySamples:         manifestInventorySamples(root, 3),
		ManifestLoopCount:                loops.Complete,
		ManifestDirectLoopCount:          loops.Direct,
		ManifestDelegatedLoopCount:       loops.Delegated,
		ManifestMissingLoopCount:         loops.Missing,
		ManifestMissingLoopByGroup:       loops.MissingGroups,
		ManifestMissingLoopSamples:       loops.MissingSamples,
		ManifestLoopBudget:               m.ManifestLoopBudget,
		UntrackedManifests:               untrackedManifests(root, m, docs),
		DirectMissingEvidenceWorkflow:    directMissingEvidenceWorkflow(root, docs),
		DirectMissingLoop:                directMissingLoops(docs), ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact, Loop: m.Loop,
	}
}

func countDocs(docs []docClass) map[string]int {
	counts := map[string]int{}
	for _, doc := range docs {
		counts[doc.Kind]++
	}
	return counts
}
