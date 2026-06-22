package main

type evidence struct {
	SchemaVersion       string       `json:"schema_version"`
	ID                  string       `json:"id"`
	Status              string       `json:"status"`
	Namespace           string       `json:"namespace"`
	Service             string       `json:"service"`
	DimensionsVerified  int          `json:"dimensions_verified"`
	JSONFieldsVerified  int          `json:"json_fields_verified"`
	ScopesVerified      int          `json:"scopes_verified"`
	MetricUnitsVerified int          `json:"metric_units_verified"`
	MetricSpecsTotal    int          `json:"metric_specs_total"`
	HTTPBreakdownRows   int          `json:"http_breakdown_rows"`
	StoreBreakdownRows  int          `json:"store_breakdown_rows"`
	SourceChecks        int          `json:"source_checks"`
	EvidenceArtifact    string       `json:"evidence_artifact"`
	Workflow            string       `json:"workflow"`
	Loop                evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, shape emfShape) evidence {
	return evidence{
		SchemaVersion:       evidenceSchema,
		ID:                  m.ID,
		Status:              "verified",
		Namespace:           shape.Namespace,
		Service:             shape.Service,
		DimensionsVerified:  len(m.RequiredDimensions),
		JSONFieldsVerified:  len(m.RequiredJSONFields),
		ScopesVerified:      len(m.RequiredScopes),
		MetricUnitsVerified: len(m.RequiredMetricUnit),
		MetricSpecsTotal:    len(shape.MetricUnits),
		HTTPBreakdownRows:   shape.HTTPBreakdownRows,
		StoreBreakdownRows:  shape.StoreBreakdownRows,
		SourceChecks:        len(m.SourceChecks),
		EvidenceArtifact:    m.EvidenceArtifact,
		Workflow:            m.Workflow,
		Loop:                m.Loop,
	}
}
