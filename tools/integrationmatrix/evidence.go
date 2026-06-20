package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	PublicGates      int          `json:"public_gates"`
	PullRequestGates int          `json:"pull_request_gates"`
	OperatorGates    int          `json:"operator_gates"`
	PrivateGates     int          `json:"private_gates"`
	WorkflowRefs     int          `json:"workflow_refs"`
	CommandRefs      int          `json:"command_refs"`
	Workflow         string       `json:"workflow"`
	GeneratedDoc     string       `json:"generated_doc"`
	Loop             evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		PublicGates:      result.PublicGates,
		PullRequestGates: result.PullRequestGates,
		OperatorGates:    result.OperatorGates,
		PrivateGates:     result.PrivateGates,
		WorkflowRefs:     result.WorkflowRefs,
		CommandRefs:      result.CommandRefs,
		Workflow:         m.Workflow,
		GeneratedDoc:     m.GeneratedDoc,
		Loop:             m.Loop,
	}
}
