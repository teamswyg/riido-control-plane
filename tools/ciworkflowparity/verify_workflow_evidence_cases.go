package main

import "strings"

func recordWorkflowEvidenceCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[9]
	record("workflow_evidence_workflow_digest_is_exact", verifyWorkflowEvidenceWorkflow(repoRoot, child))
	record("workflow_evidence_five_source_behaviors_are_native", verifyWorkflowEvidenceMapping(child))
	record("workflow_evidence_artifact_is_redacted", verifyWorkflowEvidenceArtifact(child))
	record("workflow_evidence_pipeline_preserves_order", verifyWorkflowEvidencePipeline(repoRoot, document, child))
	record("workflow_evidence_authority_remains_zero", verifyAuthorityForChild(child))
	record("workflow_evidence_classification_is_bounded", verifyWorkflowEvidenceClassification(child))
}

func verifyWorkflowEvidenceClassification(child boundedChild) bool {
	return child.Classification.Code == "workflow_evidence_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "workflow-evidence") &&
		len(child.Classification.DoesNotClaim) == 5
}
