package main

import "strings"

func recordEvidenceGraphCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[14]
	record("evidence_graph_workflow_digest_is_exact", verifyEvidenceGraphWorkflow(repoRoot, child))
	record("evidence_graph_six_source_behaviors_are_native", verifyEvidenceGraphMapping(child))
	record("evidence_graph_artifact_is_success_only_and_redacted", verifyEvidenceGraphArtifact(child))
	record("evidence_graph_pipeline_preserves_order", verifyEvidenceGraphPipeline(repoRoot, document, child))
	record("evidence_graph_authority_remains_zero", verifyAuthorityForChild(child))
	record("evidence_graph_classification_is_bounded", verifyEvidenceGraphClassification(child))
}

func verifyEvidenceGraphClassification(child boundedChild) bool {
	return child.Classification.Code == "evidence_graph_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "evidence-graph") &&
		len(child.Classification.DoesNotClaim) == 5
}
