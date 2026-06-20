package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	Boundaries       int          `json:"boundaries"`
	FocusedWorkflows int          `json:"focused_workflows"`
	SourceChecks     int          `json:"source_checks"`
	SharedContracts  int          `json:"shared_contracts"`
	RequiredPhrases  int          `json:"required_phrases"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		Boundaries:       len(m.Boundaries),
		FocusedWorkflows: len(m.FocusedWorkflows),
		SourceChecks:     countSourceChecks(m.Boundaries),
		SharedContracts:  len(m.SharedContracts),
		RequiredPhrases:  len(m.RequiredPhrases),
		Loop:             m.Loop,
	}
}

func countSourceChecks(boundaries []boundary) int {
	var count int
	for _, item := range boundaries {
		count += len(item.SourceChecks)
	}
	return count
}
