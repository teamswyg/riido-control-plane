package main

import "os"

func newManifestEvidence(m manifest, result verifyResult) manifestEvidence {
	return manifestEvidence{
		SchemaVersion: m.SchemaVersion,
		ID:            m.ID,
		Status:        "verified",
		GeneratedDoc:  m.GeneratedDoc,
		Workflow:      m.Workflow,
		WorkflowCount: result.WorkflowCount,
		PhraseChecks:  result.PhraseChecks,
		Loop:          m.Loop,
		Records:       result.Records,
	}
}

func newLiveSummary(spec workflowSpec, opt options) liveSummary {
	status := opt.LiveStatus
	if status == "" {
		status = getenvDefault("RIIDO_LIVE_CHECK_STATUS", "unknown")
	}
	return liveSummary{
		SchemaVersion:    "riido-control-plane-live-workflow-redacted-summary.v1",
		ID:               spec.ID,
		Status:           "redacted_summary",
		Workflow:         newRecord(spec),
		Run:              newRunRecord(),
		LiveStatus:       status,
		DeploymentTarget: opt.DeploymentTarget,
		Redaction:        newRedaction(spec),
	}
}

func getenvDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
