package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	WorkflowRoot     string        `json:"workflow_root"`
	AcceptedGaps     []acceptedGap `json:"accepted_gaps"`
	Assertions       []string      `json:"assertions"`
	Loop             evidenceLoop  `json:"loop"`
}

type acceptedGap struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Next   string `json:"next"`
}

type workflowRecord struct {
	Path            string `json:"path"`
	Status          string `json:"status"`
	HasExecutable   bool   `json:"has_executable"`
	HasEvidenceOut  bool   `json:"has_evidence_out"`
	UploadsArtifact bool   `json:"uploads_artifact"`
	Reason          string `json:"reason,omitempty"`
	Next            string `json:"next,omitempty"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type auditResult struct {
	Records        []workflowRecord
	Covered        int
	Accepted       int
	Unregistered   []string
	AcceptedUnused []string
}

type evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	WorkflowCount    int              `json:"workflow_count"`
	CoveredCount     int              `json:"covered_count"`
	AcceptedGapCount int              `json:"accepted_gap_count"`
	Unregistered     []string         `json:"unregistered_gaps"`
	AcceptedUnused   []string         `json:"accepted_gaps_unused"`
	Records          []workflowRecord `json:"records"`
	Loop             evidenceLoop     `json:"loop"`
}
