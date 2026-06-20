package main

type evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	CasesVerified    int            `json:"cases_verified"`
	SourceChecks     int            `json:"source_checks"`
	Results          []caseEvidence `json:"results"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	Workflow         string         `json:"workflow"`
	Loop             evidenceLoop   `json:"loop"`
}

type caseEvidence struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Records          int      `json:"records,omitempty"`
	EventTypes       []string `json:"event_types,omitempty"`
	HistoryEvents    int      `json:"history_events,omitempty"`
	AssignmentState  string   `json:"assignment_state,omitempty"`
	TaskEvents       int      `json:"task_events,omitempty"`
	OutboxErrors     int64    `json:"outbox_errors,omitempty"`
	LatencySamples   int64    `json:"latency_samples,omitempty"`
	SnapshotRestored bool     `json:"snapshot_restored,omitempty"`
}

func newEvidence(m manifest, results []caseEvidence) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		CasesVerified:    len(results),
		SourceChecks:     len(m.SourceChecks),
		Results:          results,
		EvidenceArtifact: m.EvidenceArtifact,
		Workflow:         m.Workflow,
		Loop:             m.Loop,
	}
}
