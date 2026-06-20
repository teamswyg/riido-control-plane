package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	BoundaryID       string       `json:"boundary_id,omitempty"`
	BoundaryWorkflow string       `json:"boundary_workflow,omitempty"`
	BoundaryArtifact string       `json:"boundary_evidence_artifact,omitempty"`
	Boundaries       int          `json:"boundaries"`
	FocusedWorkflows int          `json:"focused_workflows"`
	SourceChecks     int          `json:"source_checks"`
	SharedContracts  int          `json:"shared_contracts"`
	RequiredPhrases  int          `json:"required_phrases"`
	BoundaryChecks   int          `json:"boundary_source_checks,omitempty"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, selected *boundary) evidence {
	out := evidence{
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
	if selected != nil {
		out.BoundaryID = selected.ID
		out.BoundaryWorkflow = selected.Workflow
		out.BoundaryArtifact = selected.EvidenceArtifact
		out.BoundaryChecks = len(selected.SourceChecks)
	}
	return out
}

func countSourceChecks(boundaries []boundary) int {
	var count int
	for _, item := range boundaries {
		count += len(item.SourceChecks)
	}
	return count
}
