package main

type verifyResult struct {
	BoundaryCount int `json:"boundary_count"`
	EvidencePaths int `json:"evidence_paths"`
	PhraseChecks  int `json:"phrase_checks"`
	RuleCount     int `json:"rule_count"`
}

type evidence struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	BoundaryCount int        `json:"boundary_count"`
	EvidencePaths int        `json:"evidence_paths"`
	PhraseChecks  int        `json:"phrase_checks"`
	RuleCount     int        `json:"rule_count"`
	Workflow      string     `json:"workflow"`
	GeneratedDoc  string     `json:"generated_doc"`
	LinkedCD      string     `json:"linked_runtime_cd_manifest"`
	Loop          loopRecord `json:"loop"`
}

func newEvidence(m manifest, r verifyResult) evidence {
	return evidence{
		SchemaVersion: "riido-control-plane-runtime-deployment-boundary-evidence.v1",
		ID:            m.ID,
		Status:        "verified",
		BoundaryCount: r.BoundaryCount,
		EvidencePaths: r.EvidencePaths,
		PhraseChecks:  r.PhraseChecks,
		RuleCount:     r.RuleCount,
		Workflow:      m.Workflow,
		GeneratedDoc:  m.GeneratedDoc,
		LinkedCD:      m.LinkedCD,
		Loop:          m.Loop,
	}
}
