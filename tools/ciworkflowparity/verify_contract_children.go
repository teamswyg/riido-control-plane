package main

import "errors"

func verifyBoundedChildIdentities(document manifest) error {
	expected := []struct {
		id      string
		issue   string
		message string
	}{
		{"control-plane-repository-readme-ci-parity", readmeIssueURL, "repository README parity child identity drifted"},
		{"control-plane-context-map-ci-parity", contextIssueURL, "context map parity child identity drifted"},
		{"control-plane-go-ci-quality-parity", goCIIssueURL, "go CI quality parity child identity drifted"},
		{"control-plane-module-decomposition-ci-parity", moduleIssueURL, "module decomposition parity child identity drifted"},
		{"control-plane-pre-commit-baseline-ci-parity", preCommitIssueURL, "pre-commit baseline parity child identity drifted"},
		{"control-plane-migration-ledger-ci-parity", migrationIssueURL, "migration ledger parity child identity drifted"},
		{"control-plane-syntax-hash-ci-parity", syntaxHashIssueURL, "syntax hash parity child identity drifted"},
		{"control-plane-config-reference-ci-parity", configReferenceIssueURL, "config reference parity child identity drifted"},
		{"control-plane-executable-knowledge-ci-parity", executableKnowledgeIssueURL, "executable knowledge parity child identity drifted"},
		{"control-plane-workflow-evidence-ci-parity", workflowEvidenceIssueURL, "workflow evidence parity child identity drifted"},
		{"control-plane-open-questions-ci-parity", openQuestionsIssueURL, "open questions parity child identity drifted"},
		{"control-plane-harness-promotion-ci-parity", harnessPromotionIssueURL, "harness promotion parity child identity drifted"},
		{"control-plane-integration-matrix-ci-parity", integrationMatrixIssueURL, "integration matrix parity child identity drifted"},
		{"control-plane-loop-closure-audit-ci-parity", loopClosureAuditIssueURL, "loop closure audit parity child identity drifted"},
	}
	for index, want := range expected {
		child := document.BoundedChildren[index]
		if child.ID != want.id || child.Issue != want.issue || child.ParentIssue != parentIssueURL ||
			len(child.Assertions) < 5 || !completeLoop(child.Loop) {
			return errors.New(want.message)
		}
	}
	return nil
}
