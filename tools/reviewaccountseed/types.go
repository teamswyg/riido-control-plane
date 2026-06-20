package main

type manifest struct {
	SchemaVersion         string        `json:"schema_version"`
	ID                    string        `json:"id"`
	Title                 string        `json:"title"`
	GeneratedDoc          string        `json:"generated_doc"`
	Workflow              string        `json:"workflow"`
	EvidenceArtifact      string        `json:"evidence_artifact"`
	SeedSSOT              string        `json:"seed_ssot"`
	OwnerPackages         []string      `json:"owner_packages"`
	Cases                 []caseSpec    `json:"cases"`
	SourceChecks          []sourceCheck `json:"source_checks"`
	ForbiddenSeedTerms    []string      `json:"forbidden_seed_terms"`
	ForbiddenSeedTermSets []string      `json:"forbidden_seed_term_sets"`
	Loop                  evidenceLoop  `json:"loop"`
}

type caseSpec struct {
	Name                 string   `json:"name"`
	Kind                 string   `json:"kind"`
	WantTokenHashPresent bool     `json:"want_token_hash_present"`
	WantRawTokenPresent  bool     `json:"want_raw_token_present"`
	WantVisibleAgents    []string `json:"want_visible_agents"`
	WantAdmin            bool     `json:"want_admin"`
	WantProviderCount    int      `json:"want_provider_count"`
	WantAvailableCount   int      `json:"want_available_count"`
	WantChannel          string   `json:"want_channel"`
	WantCatalogStatus    int      `json:"want_catalog_status"`
	WantProviderStatus   int      `json:"want_provider_status"`
	WantPollStatus       int      `json:"want_poll_status"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
