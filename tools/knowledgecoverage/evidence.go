package main

const evidenceSchema = "riido-executable-knowledge-coverage-result.v1"

func buildEvidence(root string, m manifest, docs []docClass, problems []string) evidence {
	counts := countDocs(docs)
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
		GeneratedToolCount: generatedToolCount(docs),
		DirectSSOTCount:    counts["direct_ssot"], ManualCount: counts["manual_registered"],
		ManualByGroup: manualCountsByGroup(docs), ManualTopDirs: manualTopDirs(docs, 8),
		ManualSamples: manualSamples(docs, 2), DirectLoopCount: countDirectLoops(docs),
		GeneratedMissingTool:     generatedMissingTool(root, docs),
		GeneratedMissingWorkflow: generatedMissingWorkflow(root, docs),
		DirectMissingLoop:        directMissingLoops(docs), ProblemSummaries: problems,
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
