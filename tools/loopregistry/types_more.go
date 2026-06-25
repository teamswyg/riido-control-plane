package main

type graphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type verifyResult struct {
	Loops                     int
	Harnesses                 int
	ClosedLoops               int
	Claims                    int
	GraphEdges                int
	MaxExpiryHours            int
	Hashes                    map[string]string
	ClaimSurfaces             []claimSurface
	RefreshCadenceMinutes     map[string]int
	HarnessPromotionWorkflows map[string]string
	HarnessCandidateArtifacts map[string]string
}

type refreshPlan struct {
	LoopID               string                    `json:"loop_id"`
	Kind                 string                    `json:"kind"`
	RefreshWorkflow      string                    `json:"refresh_workflow"`
	WorkflowFile         string                    `json:"workflow_file"`
	CadenceMinutes       int                       `json:"cadence_minutes"`
	ExpiresAfterHours    int                       `json:"expires_after_hours"`
	EvidenceGeneratedAt  string                    `json:"evidence_generated_at,omitempty"`
	NextRefreshDueAt     string                    `json:"next_refresh_due_at,omitempty"`
	EvidenceExpiresAt    string                    `json:"evidence_expires_at,omitempty"`
	ManualRefreshCommand string                    `json:"manual_refresh_command"`
	EvidenceArtifacts    []string                  `json:"evidence_artifacts"`
	EvidenceRefreshes    []evidenceArtifactRefresh `json:"evidence_refreshes"`
}

type evidenceArtifactRefresh struct {
	Artifact             string `json:"artifact"`
	RefreshWorkflow      string `json:"refresh_workflow"`
	WorkflowFile         string `json:"workflow_file"`
	ManualRefreshCommand string `json:"manual_refresh_command"`
}
