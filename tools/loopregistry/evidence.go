package main

type evidence struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	LoopCount        int               `json:"loop_count"`
	HarnessCount     int               `json:"harness_count"`
	ClosedLoopCount  int               `json:"closed_loop_count"`
	ClaimCount       int               `json:"claim_count"`
	GraphEdgeCount   int               `json:"graph_edge_count"`
	MaxExpiryHours   int               `json:"max_expiry_hours"`
	SemanticHashes   map[string]string `json:"semantic_hashes"`
	Workflow         string            `json:"workflow"`
	GeneratedDoc     string            `json:"generated_doc"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	Loop             evidenceLoop      `json:"loop"`
	Impact           *impactEvidence   `json:"impact,omitempty"`
}

func newEvidence(m manifest, result verifyResult, impact *impactEvidence) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		LoopCount:        result.Loops,
		HarnessCount:     result.Harnesses,
		ClosedLoopCount:  result.ClosedLoops,
		ClaimCount:       result.Claims,
		GraphEdgeCount:   result.GraphEdges,
		MaxExpiryHours:   result.MaxExpiryHours,
		SemanticHashes:   result.Hashes,
		Workflow:         m.Workflow,
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.EvidenceArtifact,
		Loop:             m.Loop,
		Impact:           impact,
	}
}
