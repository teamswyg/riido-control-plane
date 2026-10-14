package main

import "strings"

func recordOpenQuestionsCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[10]
	record("open_questions_workflow_digest_is_exact", verifyOpenQuestionsWorkflow(repoRoot, child))
	record("open_questions_six_source_behaviors_are_native", verifyOpenQuestionsMapping(child))
	record("open_questions_artifact_is_redacted", verifyOpenQuestionsArtifact(child))
	record("open_questions_pipeline_preserves_order", verifyOpenQuestionsPipeline(repoRoot, document, child))
	record("open_questions_authority_remains_zero", verifyAuthorityForChild(child))
	record("open_questions_classification_is_bounded", verifyOpenQuestionsClassification(child))
}

func verifyOpenQuestionsClassification(child boundedChild) bool {
	return child.Classification.Code == "open_questions_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "open-questions") &&
		len(child.Classification.DoesNotClaim) == 5
}
