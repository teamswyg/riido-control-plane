package main

type verifyResult struct {
	QuestionCount int            `json:"question_count"`
	OpenCount     int            `json:"open_count"`
	ResolvedCount int            `json:"resolved_count"`
	StatusCounts  map[string]int `json:"status_counts"`
}

type evidence struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Status        string            `json:"status"`
	Result        verifyResult      `json:"result"`
	OpenCommands  []questionCommand `json:"open_commands"`
	Workflow      string            `json:"workflow"`
	GeneratedDoc  string            `json:"generated_doc"`
	Loop          loopRecord        `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion: "riido-control-plane-open-questions-evidence.v1",
		ID:            m.ID,
		Status:        "verified",
		Result:        result,
		OpenCommands:  openQuestionCommands(m.Questions),
		Workflow:      m.Workflow,
		GeneratedDoc:  m.GeneratedDoc,
		Loop:          m.Loop,
	}
}
