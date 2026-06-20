package main

type evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	CasesVerified    int            `json:"cases_verified"`
	SourceChecks     int            `json:"source_checks"`
	SeedSSOT         string         `json:"seed_ssot"`
	Results          []caseEvidence `json:"results"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	Workflow         string         `json:"workflow"`
	Loop             evidenceLoop   `json:"loop"`
}

type caseEvidence struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	TokenHashPresent bool     `json:"token_hash_present,omitempty"`
	RawTokenPresent  bool     `json:"raw_token_present"`
	VisibleAgents    []string `json:"visible_agents,omitempty"`
	Admin            bool     `json:"admin,omitempty"`
	ProviderCount    int      `json:"provider_count,omitempty"`
	AvailableCount   int      `json:"available_count"`
	Channel          string   `json:"channel,omitempty"`
	CatalogStatus    int      `json:"catalog_status,omitempty"`
	ProviderStatus   int      `json:"provider_status,omitempty"`
	PollStatus       int      `json:"poll_status,omitempty"`
}

func newEvidence(m manifest, results []caseEvidence) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		CasesVerified:    len(results),
		SourceChecks:     len(m.SourceChecks),
		SeedSSOT:         m.SeedSSOT,
		Results:          results,
		EvidenceArtifact: m.EvidenceArtifact,
		Workflow:         m.Workflow,
		Loop:             m.Loop,
	}
}
