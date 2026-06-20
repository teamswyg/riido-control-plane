package main

type evidence struct {
	SchemaVersion       string       `json:"schema_version"`
	ID                  string       `json:"id"`
	Status              string       `json:"status"`
	Endpoint            string       `json:"endpoint"`
	AuthorizedStatus    int          `json:"authorized_status"`
	MissingScopeStatus  int          `json:"missing_scope_status"`
	UnconfiguredStatus  int          `json:"store_unconfigured_status"`
	MetricsSchema       string       `json:"metrics_schema"`
	JSONFieldsVerified  int          `json:"json_fields_verified"`
	StatusCasesVerified int          `json:"status_cases_verified"`
	HTTPBreakdownRows   int          `json:"http_breakdown_rows"`
	StoreBreakdownRows  int          `json:"store_breakdown_rows"`
	SourceChecks        int          `json:"source_checks"`
	EvidenceArtifact    string       `json:"evidence_artifact"`
	Workflow            string       `json:"workflow"`
	Loop                evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, result adapterResult) evidence {
	return evidence{
		SchemaVersion:       evidenceSchema,
		ID:                  m.ID,
		Status:              "verified",
		Endpoint:            m.Endpoint.Method + " " + m.Endpoint.Path,
		AuthorizedStatus:    result.AuthorizedStatus,
		MissingScopeStatus:  result.MissingScopeStatus,
		UnconfiguredStatus:  result.UnconfiguredStatus,
		MetricsSchema:       result.SchemaVersion,
		JSONFieldsVerified:  len(m.RequiredFields),
		StatusCasesVerified: len(m.RequiredStatuses),
		HTTPBreakdownRows:   result.HTTPBreakdownRows,
		StoreBreakdownRows:  result.StoreBreakdownRows,
		SourceChecks:        len(m.SourceChecks),
		EvidenceArtifact:    m.EvidenceArtifact,
		Workflow:            m.Workflow,
		Loop:                m.Loop,
	}
}
