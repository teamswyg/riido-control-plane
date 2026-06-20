package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	OwnedContexts    int          `json:"owned_contexts"`
	ImportedContexts int          `json:"imported_contexts"`
	ExternalContexts int          `json:"external_contexts"`
	SSOTLinks        int          `json:"ssot_links"`
	SourceChecks     int          `json:"source_checks"`
	ForbiddenImports int          `json:"forbidden_imports_found"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		OwnedContexts:    len(m.OwnedContexts),
		ImportedContexts: len(m.ImportedContexts),
		ExternalContexts: len(m.ExternalContexts),
		SSOTLinks:        len(m.SSOTLinks),
		SourceChecks:     len(m.SourceChecks),
		ForbiddenImports: 0,
		Loop:             m.Loop,
	}
}
