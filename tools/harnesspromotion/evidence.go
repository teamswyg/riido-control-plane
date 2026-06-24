package main

type evidence struct {
	SchemaVersion        string            `json:"schema_version"`
	ID                   string            `json:"id"`
	Status               string            `json:"status"`
	SourceCount          int               `json:"source_count"`
	RequiredRefs         int               `json:"required_refs"`
	Workflow             string            `json:"workflow"`
	GeneratedDoc         string            `json:"generated_doc"`
	EvidenceArtifact     string            `json:"evidence_artifact"`
	LoopRegistryManifest string            `json:"loop_registry_manifest"`
	Loop                 evidenceLoop      `json:"loop"`
	Sources              []promotionSource `json:"sources"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:        evidenceSchema,
		ID:                   m.ID,
		Status:               "verified",
		SourceCount:          result.SourceCount,
		RequiredRefs:         result.ClaimCount,
		Workflow:             m.Workflow,
		GeneratedDoc:         m.GeneratedDoc,
		EvidenceArtifact:     m.EvidenceArtifact,
		LoopRegistryManifest: m.LoopRegistryManifest,
		Loop:                 m.Loop,
		Sources:              m.Sources,
	}
}
