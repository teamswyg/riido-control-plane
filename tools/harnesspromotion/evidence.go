package main

type evidence struct {
	SchemaVersion                   string            `json:"schema_version"`
	ID                              string            `json:"id"`
	Status                          string            `json:"status"`
	GeneratedAt                     string            `json:"generated_at"`
	ExpiresAt                       string            `json:"expires_at"`
	SourceCount                     int               `json:"source_count"`
	SidecarSourceCount              int               `json:"sidecar_source_count"`
	LoopOwnedCandidateProducerCount int               `json:"loop_owned_candidate_producer_count"`
	RequiredRefs                    int               `json:"required_refs"`
	Workflow                        string            `json:"workflow"`
	GeneratedDoc                    string            `json:"generated_doc"`
	EvidenceArtifact                string            `json:"evidence_artifact"`
	LoopRegistryManifest            string            `json:"loop_registry_manifest"`
	PromotionEdgeCount              int               `json:"promotion_edge_count"`
	PromotionEdges                  []graphEdge       `json:"promotion_edges"`
	PromotionResult                 *promotionResult  `json:"promotion_result,omitempty"`
	Loop                            evidenceLoop      `json:"loop"`
	Sources                         []promotionSource `json:"sources"`
}

func newEvidence(m manifest, result verifyResult, promotion *promotionResult) evidence {
	generatedAt, expiresAt := evidenceWindow(harnessPromotionEvidenceTTLHours)
	edges := promotionEdgesForSources(m.Sources)
	return evidence{
		SchemaVersion:                   evidenceSchema,
		ID:                              m.ID,
		Status:                          "verified",
		GeneratedAt:                     generatedAt,
		ExpiresAt:                       expiresAt,
		SourceCount:                     result.SourceCount,
		SidecarSourceCount:              result.SidecarSourceCount,
		LoopOwnedCandidateProducerCount: result.LoopOwnedCandidateProducerCount,
		RequiredRefs:                    result.ClaimCount,
		Workflow:                        m.Workflow,
		GeneratedDoc:                    m.GeneratedDoc,
		EvidenceArtifact:                m.EvidenceArtifact,
		LoopRegistryManifest:            m.LoopRegistryManifest,
		PromotionEdgeCount:              len(edges),
		PromotionEdges:                  edges,
		PromotionResult:                 promotion,
		Loop:                            m.Loop,
		Sources:                         m.Sources,
	}
}
