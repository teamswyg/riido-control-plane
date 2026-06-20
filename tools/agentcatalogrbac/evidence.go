package main

type evidence struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Rules         int          `json:"visibility_rules"`
	Scopes        int          `json:"authorization_scopes"`
	Routes        int          `json:"routes"`
	SourceChecks  int          `json:"source_checks"`
	Loop          evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion: evidenceSchema,
		ID:            m.ID,
		Status:        "verified",
		Rules:         len(m.VisibilityRules),
		Scopes:        len(m.AuthorizationScopes),
		Routes:        len(m.Routes),
		SourceChecks:  len(m.SourceChecks),
		Loop:          m.Loop,
	}
}
