package main

type evidence struct {
	SchemaVersion             string                         `json:"schema_version"`
	ID                        string                         `json:"id"`
	Status                    string                         `json:"status"`
	GeneratedAt               string                         `json:"generated_at"`
	ExpiresAt                 string                         `json:"expires_at"`
	LoopCount                 int                            `json:"loop_count"`
	HarnessCount              int                            `json:"harness_count"`
	ClosedLoopCount           int                            `json:"closed_loop_count"`
	ClaimCount                int                            `json:"claim_count"`
	GraphEdgeCount            int                            `json:"graph_edge_count"`
	MaxExpiryHours            int                            `json:"max_expiry_hours"`
	SemanticHashes            map[string]string              `json:"semantic_hashes"`
	EvidenceKinds             []evidenceKind                 `json:"evidence_kinds"`
	LoopSurfaces              []loopSurface                  `json:"loop_surfaces"`
	LoopCompletionSummary     loopCompletionSummary          `json:"loop_completion_summary"`
	LoopCompletions           []loopCompletion               `json:"loop_completions"`
	CoverageDimensions        []loopCoverageDimensionSurface `json:"coverage_dimensions"`
	ArchitectureIndex         architectureIndex              `json:"architecture_index"`
	EvidenceGraph             []graphEdge                    `json:"evidence_graph"`
	ClaimSurfaces             []claimSurface                 `json:"claim_surfaces"`
	RefreshWorkflows          map[string]string              `json:"refresh_workflows"`
	RefreshCadenceMinutes     map[string]int                 `json:"refresh_cadence_minutes"`
	RefreshPlanSummary        refreshPlanSummary             `json:"refresh_plan_summary"`
	RefreshPlans              []refreshPlan                  `json:"refresh_plans"`
	ProviderCoverage          map[string][]string            `json:"provider_coverage,omitempty"`
	HarnessWorkflowExclusions []harnessWorkflowExclusion     `json:"harness_workflow_exclusions,omitempty"`
	HarnessPromotionWorkflows map[string]string              `json:"harness_promotion_workflows"`
	HarnessCandidateArtifacts map[string]string              `json:"harness_candidate_artifacts"`
	Workflow                  string                         `json:"workflow"`
	GeneratedDoc              string                         `json:"generated_doc"`
	EvidenceArtifact          string                         `json:"evidence_artifact"`
	Loop                      evidenceLoop                   `json:"loop"`
	Impact                    *impactEvidence                `json:"impact,omitempty"`
}
