package main

import "strings"

func recordLoopClosureAuditCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[13]
	record("loop_closure_audit_workflow_digest_is_exact", verifyLoopClosureAuditWorkflow(repoRoot, child))
	record("loop_closure_audit_six_source_behaviors_are_native", verifyLoopClosureAuditMapping(child))
	record("loop_closure_audit_artifacts_are_redacted", verifyLoopClosureAuditArtifact(child))
	record("loop_closure_audit_pipeline_preserves_order", verifyLoopClosureAuditPipeline(repoRoot, document, child))
	record("loop_closure_audit_authority_remains_zero", verifyAuthorityForChild(child))
	record("loop_closure_audit_classification_is_bounded", verifyLoopClosureAuditClassification(child))
}

func verifyLoopClosureAuditClassification(child boundedChild) bool {
	return child.Classification.Code == "loop_closure_audit_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "loop-closure-audit") &&
		len(child.Classification.DoesNotClaim) == 5
}
