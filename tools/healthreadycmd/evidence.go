package main

type evidence struct {
	SchemaVersion       string             `json:"schema_version"`
	ID                  string             `json:"id"`
	Status              string             `json:"status"`
	EndpointsVerified   int                `json:"endpoints_verified"`
	SourceChecks        int                `json:"source_checks"`
	CommandTestsAnchors int                `json:"command_test_anchors"`
	Results             []endpointEvidence `json:"results"`
	EvidenceArtifact    string             `json:"evidence_artifact"`
	Workflow            string             `json:"workflow"`
	Loop                evidenceLoop       `json:"loop"`
}

type endpointEvidence struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	HTTPStatus int    `json:"http_status"`
	Status     string `json:"status"`
}

func newEvidence(m manifest, results []endpointEvidence) evidence {
	return evidence{
		SchemaVersion:       evidenceSchema,
		ID:                  m.ID,
		Status:              "verified",
		EndpointsVerified:   len(results),
		SourceChecks:        len(m.SourceChecks),
		CommandTestsAnchors: len(m.CommandTests),
		Results:             results,
		EvidenceArtifact:    m.EvidenceArtifact,
		Workflow:            m.Workflow,
		Loop:                m.Loop,
	}
}
