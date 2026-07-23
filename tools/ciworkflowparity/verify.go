package main

import "errors"

func verify(repoRoot, contractPath string) (evidence, error) {
	document, err := loadManifest(repoRoot, contractPath)
	if err != nil {
		return evidence{}, err
	}
	cases := map[string]string{}
	record := func(name string, passed bool) { cases[name] = status(passed) }
	record("legacy_workflow_digest_is_exact", verifyLegacyWorkflow(repoRoot, document))
	record("go_toolchain_is_bound_to_go_mod", verifyGoToolchain(repoRoot, document))
	record("all_five_source_steps_are_native", verifyNativeMapping(document))
	record("native_artifact_preserves_failure_evidence", verifyArtifactMapping(document))
	record("native_pipeline_preserves_step_order", verifyPipeline(repoRoot, document))
	record("runner_is_pinned_and_credential_free", verifyRunner(repoRoot, document))
	record("retirement_and_runtime_authority_remain_zero", verifyAuthority(document))
	record("rollback_preserves_exact_baseline", verifyRollback(document))
	record("classification_is_source_complete_not_retired", verifyClassification(document))
	child := document.BoundedChildren[0]
	record("repository_readme_workflow_digest_is_exact", verifyReadmeWorkflow(repoRoot, child))
	record("repository_readme_steps_are_native", verifyReadmeMapping(child))
	record("repository_readme_artifact_is_redacted", verifyReadmeArtifact(child))
	record("repository_readme_pipeline_preserves_order", verifyReadmePipeline(repoRoot, document, child))
	record("repository_readme_authority_remains_zero", verifyAuthorityForChild(child))
	record("repository_readme_classification_is_bounded", verifyReadmeClassification(child))
	recordContextMapCases(repoRoot, document, record)
	recordGoCICases(repoRoot, document, record)
	recordModuleDecompositionCases(repoRoot, document, record)
	recordPreCommitCases(repoRoot, document, record)
	recordMigrationLedgerCases(repoRoot, document, record)
	recordSyntaxHashCases(repoRoot, document, record)
	recordConfigReferenceCases(repoRoot, document, record)
	recordExecutableKnowledgeCases(repoRoot, document, record)
	recordWorkflowEvidenceCases(repoRoot, document, record)
	recordOpenQuestionsCases(repoRoot, document, record)
	recordHarnessPromotionCases(repoRoot, document, record)
	result := newEvidence(document, cases)
	if result.Decision != "passed" {
		return result, errors.New("control plane baseline CI parity evidence failed closed")
	}
	return result, nil
}

func newEvidence(document manifest, cases map[string]string) evidence {
	result := evidence{
		SchemaVersion: evidenceSchema, Decision: "passed", Cases: cases,
		BaselineWorkflowSHA256: document.Baseline.WorkflowSHA256,
		PipelineID:             document.Runner.PipelineID, RunnerRevision: document.Runner.Revision,
		RequiredAdapterCount:    document.ParityClaim.RequiredAdapterCount,
		LegacyWorkflowPreserved: document.Rollback.BaselineWorkflowPreserved && !document.ParityClaim.SourceWorkflowEdited,
		RetirementAuthorized:    document.Authority.WorkflowRetirementAuthorized,
		RuntimeEffect:           document.Authority.RuntimeEffect,
		BoundedChildren:         []childEvidence{},
	}
	for _, child := range document.BoundedChildren {
		result.BoundedChildren = append(result.BoundedChildren, childEvidence{
			ID: child.ID, Issue: child.Issue, WorkflowSHA256: child.Baseline.WorkflowSHA256,
			RequiredAdapterCount:    child.ParityClaim.RequiredAdapterCount,
			LegacyWorkflowPreserved: child.Rollback.BaselineWorkflowPreserved && !child.ParityClaim.SourceWorkflowEdited,
			RetirementAuthorized:    child.Authority.WorkflowRetirementAuthorized,
			RuntimeEffect:           child.Authority.RuntimeEffect,
		})
	}
	for _, value := range cases {
		if value != "passed" {
			result.Decision = "failed"
		}
	}
	return result
}
