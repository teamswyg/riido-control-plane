package main

type evidence struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	GeneratedAt      string            `json:"generated_at"`
	ExpiresAt        string            `json:"expires_at"`
	ChainCount       int               `json:"chain_count"`
	ClaimRefs        int               `json:"claim_ref_count"`
	ChangeRefs       int               `json:"change_ref_count"`
	VerifierRefs     int               `json:"verifier_ref_count"`
	EvidenceRefs     int               `json:"evidence_ref_count"`
	Workflow         string            `json:"workflow"`
	WorkflowFile     string            `json:"workflow_file"`
	CadenceMinutes   int               `json:"cadence_minutes"`
	ManualRefresh    string            `json:"manual_refresh_command"`
	GeneratedDoc     string            `json:"generated_doc"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	LoopRegistry     string            `json:"loop_registry_manifest"`
	Loop             loopRecord        `json:"loop"`
	ChainSummary     chainSummary      `json:"chain_summary"`
	NextLoopSummary  []nextLoopSummary `json:"next_loop_summary"`
	Chains           []chainEvidence   `json:"chains"`
	Impact           *impactEvidence   `json:"impact,omitempty"`
}

func newEvidence(m manifest, result verifyResult, impact *impactEvidence) evidence {
	generatedAt, expiresAt := evidenceWindow(evidenceGraphEvidenceTTLHours)
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		GeneratedAt:      generatedAt,
		ExpiresAt:        expiresAt,
		ChainCount:       result.Chains,
		ClaimRefs:        result.ClaimRefs,
		ChangeRefs:       result.ChangeRefs,
		VerifierRefs:     result.VerifierRefs,
		EvidenceRefs:     result.EvidenceRefs,
		Workflow:         m.Workflow,
		WorkflowFile:     workflowFile(m.Workflow),
		CadenceMinutes:   result.RefreshCadenceMinutes,
		ManualRefresh:    manualRefreshCommand(m.Workflow),
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.Evidence,
		LoopRegistry:     m.LoopRegistry,
		Loop:             m.Loop,
		ChainSummary:     summarizeChains(m.Chains),
		NextLoopSummary:  summarizeNextLoops(m.Chains),
		Chains:           evidenceChains(m.Chains),
		Impact:           impact,
	}
}
