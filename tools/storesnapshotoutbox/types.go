package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	OwnerPackage     string        `json:"owner_package"`
	Cases            []caseSpec    `json:"cases"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Loop             evidenceLoop  `json:"loop"`
}

type caseSpec struct {
	Name                string   `json:"name"`
	Kind                string   `json:"kind"`
	WantRecords         int      `json:"want_records"`
	WantEvents          []string `json:"want_events"`
	WantHistoryEvents   int      `json:"want_history_events"`
	WantAssignmentState string   `json:"want_assignment_state"`
	WantTaskEvents      int      `json:"want_task_events"`
	WantOutboxErrors    int64    `json:"want_outbox_errors"`
	WantLatencySamples  int64    `json:"want_latency_samples"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
