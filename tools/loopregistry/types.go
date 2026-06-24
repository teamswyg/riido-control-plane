package main

type manifest struct {
	SchemaVersion             string                     `json:"schema_version"`
	ID                        string                     `json:"id"`
	Title                     string                     `json:"title"`
	GeneratedDoc              string                     `json:"generated_doc"`
	Workflow                  string                     `json:"workflow"`
	EvidenceArtifact          string                     `json:"evidence_artifact"`
	EvidenceTool              string                     `json:"evidence_tool"`
	Assertions                []string                   `json:"assertions"`
	Loops                     []loopRecord               `json:"loops"`
	HarnessWorkflowExclusions []harnessWorkflowExclusion `json:"harness_workflow_exclusions,omitempty"`
	Claims                    []claimBinding             `json:"claim_bindings"`
	EvidenceGraph             []graphEdge                `json:"evidence_graph"`
	Loop                      evidenceLoop               `json:"loop"`
}

type loopRecord struct {
	ID                string           `json:"id"`
	Kind              string           `json:"kind"`
	Observes          []string         `json:"observes"`
	Verifies          []string         `json:"verifies"`
	Evidence          []evidenceSource `json:"evidence"`
	RefreshWorkflow   string           `json:"refresh_workflow"`
	ExpiresAfterHours int              `json:"expires_after_hours"`
	FailsWhen         []string         `json:"fails_when"`
	PromotesTo        []string         `json:"promotes_to,omitempty"`
	Providers         []string         `json:"providers,omitempty"`
}

type evidenceSource struct {
	Kind            string `json:"kind"`
	Path            string `json:"path"`
	Redacted        bool   `json:"redacted,omitempty"`
	RefreshWorkflow string `json:"refresh_workflow,omitempty"`
}

type harnessWorkflowExclusion struct {
	Workflow string `json:"workflow"`
	Reason   string `json:"reason"`
}
