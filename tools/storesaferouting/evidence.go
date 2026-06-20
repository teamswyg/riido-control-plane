package main

type evidence struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	CasesVerified    int               `json:"cases_verified"`
	SourceChecks     int               `json:"source_checks"`
	DomainSSOT       string            `json:"domain_ssot"`
	Results          []routingEvidence `json:"results"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	Workflow         string            `json:"workflow"`
	Loop             evidenceLoop      `json:"loop"`
}

type routingEvidence struct {
	Name          string `json:"name"`
	Runtime       string `json:"runtime_provider"`
	Allowed       bool   `json:"allowed"`
	RoutingStatus string `json:"routing_status,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Error         string `json:"error,omitempty"`
}

func newEvidence(m manifest, results []routingEvidence) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		CasesVerified:    len(results),
		SourceChecks:     len(m.SourceChecks),
		DomainSSOT:       m.DomainSSOT,
		Results:          results,
		EvidenceArtifact: m.EvidenceArtifact,
		Workflow:         m.Workflow,
		Loop:             m.Loop,
	}
}
