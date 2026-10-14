package main

type evidence struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Workflow      string     `json:"workflow"`
	GeneratedDoc  string     `json:"generated_doc"`
	GateCount     int        `json:"gate_count"`
	PhraseChecks  int        `json:"phrase_checks"`
	Loop          loopRecord `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion: evidenceSchema,
		ID:            m.ID,
		Status:        "verified",
		Workflow:      m.Workflow,
		GeneratedDoc:  m.GeneratedDoc,
		GateCount:     result.Gates,
		PhraseChecks:  result.PhraseChecks,
		Loop:          m.Loop,
	}
}
