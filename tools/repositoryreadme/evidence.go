package main

type evidence struct {
	SchemaVersion       string       `json:"schema_version"`
	ID                  string       `json:"id"`
	Status              string       `json:"status"`
	Manifest            string       `json:"manifest"`
	GeneratedDoc        string       `json:"generated_doc"`
	FragmentCount       int          `json:"fragment_count"`
	DocLinkCount        int          `json:"doc_link_count"`
	EndpointCount       int          `json:"endpoint_count"`
	VerificationCount   int          `json:"verification_count"`
	RuntimeCDNoteCount  int          `json:"runtime_cd_note_count"`
	RequiredMarkerCount int          `json:"required_marker_count"`
	Loop                evidenceLoop `json:"loop"`
}

func buildEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion: evidenceSchema, ID: m.ID, Status: "verified",
		Manifest: defaultManifest, GeneratedDoc: generatedDoc, FragmentCount: len(m.Fragments),
		DocLinkCount: len(m.DocLinks), EndpointCount: len(m.Development.Endpoints),
		VerificationCount: len(m.Verification), RuntimeCDNoteCount: len(m.RuntimeCD.Notes),
		RequiredMarkerCount: len(m.RequiredMarkers), Loop: m.Loop,
	}
}
