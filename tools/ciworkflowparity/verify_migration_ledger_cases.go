package main

import "strings"

func recordMigrationLedgerCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[5]
	record("migration_ledger_workflow_digest_is_exact", verifyMigrationLedgerWorkflow(repoRoot, child))
	record("migration_ledger_six_source_behaviors_are_native", verifyMigrationLedgerMapping(child))
	record("migration_ledger_artifact_is_redacted", verifyMigrationLedgerArtifact(child))
	record("migration_ledger_pipeline_preserves_order", verifyMigrationLedgerPipeline(repoRoot, document, child))
	record("migration_ledger_authority_remains_zero", verifyAuthorityForChild(child))
	record("migration_ledger_classification_is_bounded", verifyMigrationLedgerClassification(child))
}

func verifyMigrationLedgerClassification(child boundedChild) bool {
	return child.Classification.Code == "migration_ledger_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "migration-ledger") &&
		len(child.Classification.DoesNotClaim) == 5
}
