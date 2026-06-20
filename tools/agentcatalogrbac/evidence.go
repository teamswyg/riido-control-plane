package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	ProfileID        string       `json:"profile_id"`
	Workflow         string       `json:"workflow"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	Focus            string       `json:"focus"`
	TestPatterns     []string     `json:"test_patterns"`
	Rules            int          `json:"visibility_rules"`
	Scopes           int          `json:"authorization_scopes"`
	Routes           int          `json:"routes"`
	RequestDTOs      int          `json:"request_dtos"`
	ResponseDTOs     int          `json:"response_dtos"`
	StoreMethods     int          `json:"store_methods"`
	SourceChecks     int          `json:"source_checks"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, p profile) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		ProfileID:        p.ID,
		Workflow:         p.Workflow,
		EvidenceArtifact: p.EvidenceArtifact,
		Focus:            p.Focus,
		TestPatterns:     append([]string{}, p.TestPatterns...),
		Rules:            len(m.VisibilityRules),
		Scopes:           len(m.AuthorizationScopes),
		Routes:           len(m.Routes),
		RequestDTOs:      len(m.RequestDTOs),
		ResponseDTOs:     len(m.ResponseDTOs),
		StoreMethods:     len(m.StoreMethods),
		SourceChecks:     len(m.SourceChecks),
		Loop:             m.Loop,
	}
}
