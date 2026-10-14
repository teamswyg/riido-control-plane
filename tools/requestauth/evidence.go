package main

type evidence struct {
	SchemaVersion    string `json:"schema_version"`
	ID               string `json:"id"`
	Status           string `json:"status"`
	Surfaces         int    `json:"surfaces"`
	Resources        int    `json:"resources"`
	TokenTransports  int    `json:"token_transports"`
	RuntimeConfigs   int    `json:"runtime_configs"`
	ContractVersions int    `json:"contract_versions"`
	RuleGroups       int    `json:"rule_groups"`
	Rules            int    `json:"rules"`
	SourceChecks     int    `json:"source_checks"`
	Loop             loop   `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		Surfaces:         len(m.Surfaces),
		Resources:        len(m.Resources),
		TokenTransports:  len(m.TokenTransports),
		RuntimeConfigs:   len(m.RuntimeConfigKeys),
		ContractVersions: len(m.ExternalContractVersions),
		RuleGroups:       len(m.RuleGroups),
		Rules:            countRules(m.RuleGroups),
		SourceChecks:     len(m.SourceChecks),
		Loop:             m.Loop,
	}
}

func countRules(groups []ruleGroup) int {
	total := 0
	for _, group := range groups {
		total += len(group.Rules)
	}
	return total
}
