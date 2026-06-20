package main

type evidence struct {
	SchemaVersion     string         `json:"schema_version"`
	ID                string         `json:"id"`
	Status            string         `json:"status"`
	CasesVerified     int            `json:"cases_verified"`
	SourceChecks      int            `json:"source_checks"`
	RuntimeConfigKeys []string       `json:"runtime_config_keys"`
	Results           []corsEvidence `json:"results"`
	EvidenceArtifact  string         `json:"evidence_artifact"`
	Workflow          string         `json:"workflow"`
	Loop              evidenceLoop   `json:"loop"`
}

type corsEvidence struct {
	Name        string `json:"name"`
	Route       string `json:"route"`
	Origin      string `json:"origin"`
	HTTPStatus  int    `json:"http_status"`
	AllowOrigin string `json:"allow_origin"`
	Credentials string `json:"credentials"`
}

func newEvidence(m manifest, results []corsEvidence) evidence {
	return evidence{
		SchemaVersion:     evidenceSchema,
		ID:                m.ID,
		Status:            "verified",
		CasesVerified:     len(results),
		SourceChecks:      len(m.SourceChecks),
		RuntimeConfigKeys: m.RuntimeConfigKeys,
		Results:           results,
		EvidenceArtifact:  m.EvidenceArtifact,
		Workflow:          m.Workflow,
		Loop:              m.Loop,
	}
}
