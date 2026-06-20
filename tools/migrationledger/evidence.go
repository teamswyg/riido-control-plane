package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	SectionCount     int          `json:"section_count"`
	SliceCount       int          `json:"slice_count"`
	ValidationGates  int          `json:"validation_gates"`
	RiidoReferences  int          `json:"riido_references"`
	GeneratedDoc     string       `json:"generated_doc"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		SectionCount:     result.Sections,
		SliceCount:       result.Slices,
		ValidationGates:  result.ValidationGates,
		RiidoReferences:  result.RiidoReferences,
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.EvidenceArtifact,
		Loop:             m.Loop,
	}
}
