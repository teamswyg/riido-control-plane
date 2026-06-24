package main

type evidence struct {
	SchemaVersion             string            `json:"schema_version"`
	ID                        string            `json:"id"`
	Status                    string            `json:"status"`
	GeneratedAt               string            `json:"generated_at"`
	ExpiresAt                 string            `json:"expires_at"`
	LoopCount                 int               `json:"loop_count"`
	HarnessCount              int               `json:"harness_count"`
	ClosedLoopCount           int               `json:"closed_loop_count"`
	ClaimCount                int               `json:"claim_count"`
	GraphEdgeCount            int               `json:"graph_edge_count"`
	MaxExpiryHours            int               `json:"max_expiry_hours"`
	SemanticHashes            map[string]string `json:"semantic_hashes"`
	ClaimSurfaces             []claimSurface    `json:"claim_surfaces"`
	RefreshWorkflows          map[string]string `json:"refresh_workflows"`
	RefreshCadenceMinutes     map[string]int    `json:"refresh_cadence_minutes"`
	RefreshPlans              []refreshPlan     `json:"refresh_plans"`
	HarnessPromotionWorkflows map[string]string `json:"harness_promotion_workflows"`
	HarnessCandidateArtifacts map[string]string `json:"harness_candidate_artifacts"`
	Workflow                  string            `json:"workflow"`
	GeneratedDoc              string            `json:"generated_doc"`
	EvidenceArtifact          string            `json:"evidence_artifact"`
	Loop                      evidenceLoop      `json:"loop"`
	Impact                    *impactEvidence   `json:"impact,omitempty"`
}

func newEvidence(m manifest, result verifyResult, impact *impactEvidence) evidence {
	generatedAt, expiresAt := evidenceWindow(loopRegistryEvidenceTTLHours)
	return evidence{
		SchemaVersion:             evidenceSchema,
		ID:                        m.ID,
		Status:                    "verified",
		GeneratedAt:               generatedAt,
		ExpiresAt:                 expiresAt,
		LoopCount:                 result.Loops,
		HarnessCount:              result.Harnesses,
		ClosedLoopCount:           result.ClosedLoops,
		ClaimCount:                result.Claims,
		GraphEdgeCount:            result.GraphEdges,
		MaxExpiryHours:            result.MaxExpiryHours,
		SemanticHashes:            result.Hashes,
		ClaimSurfaces:             result.ClaimSurfaces,
		RefreshWorkflows:          refreshWorkflows(m.Loops),
		RefreshCadenceMinutes:     result.RefreshCadenceMinutes,
		RefreshPlans:              refreshPlans(m.Loops, result.RefreshCadenceMinutes),
		HarnessPromotionWorkflows: result.HarnessPromotionWorkflows,
		HarnessCandidateArtifacts: result.HarnessCandidateArtifacts,
		Workflow:                  m.Workflow,
		GeneratedDoc:              m.GeneratedDoc,
		EvidenceArtifact:          m.EvidenceArtifact,
		Loop:                      m.Loop,
		Impact:                    impact,
	}
}

func refreshWorkflows(loops []loopRecord) map[string]string {
	out := map[string]string{}
	for _, loop := range loops {
		out[loop.ID] = loop.RefreshWorkflow
	}
	return out
}
