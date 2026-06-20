package main

type evidence struct {
	SchemaVersion        string       `json:"schema_version"`
	ID                   string       `json:"id"`
	Status               string       `json:"status"`
	Surfaces             int          `json:"surfaces"`
	RoutingStatuses      int          `json:"routing_statuses"`
	DistributionChannels int          `json:"distribution_channels"`
	ValidationRules      int          `json:"validation_rules"`
	RoutingRules         int          `json:"routing_rules"`
	AuthorizationRules   int          `json:"authorization_rules"`
	SourceChecks         int          `json:"source_checks"`
	Loop                 evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:        evidenceSchema,
		ID:                   m.ID,
		Status:               "verified",
		Surfaces:             len(m.Surfaces),
		RoutingStatuses:      len(m.RoutingStatuses),
		DistributionChannels: len(m.DistributionChannels),
		ValidationRules:      len(m.ValidationRules),
		RoutingRules:         len(m.RoutingRules),
		AuthorizationRules:   len(m.Authorization),
		SourceChecks:         len(m.SourceChecks),
		Loop:                 m.Loop,
	}
}
