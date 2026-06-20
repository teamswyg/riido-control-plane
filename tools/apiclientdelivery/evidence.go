package main

type verifyResult struct {
	SourceManifests int            `json:"source_manifests"`
	Owners          int            `json:"owners"`
	FigmaContexts   int            `json:"figma_contexts"`
	SourceChecks    int            `json:"source_checks"`
	PhraseChecks    int            `json:"phrase_checks"`
	ForbiddenChecks int            `json:"forbidden_checks"`
	RiskTests       int            `json:"risk_tests"`
	RiskEvidence    []riskEvidence `json:"-"`
}

type evidence struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Result        verifyResult `json:"result"`
	Workflow      string       `json:"workflow"`
	GeneratedDoc  string       `json:"generated_doc"`
	Loop          loopRecord   `json:"loop"`
}

func newEvidence(m manifest, r verifyResult) evidence {
	return evidence{
		SchemaVersion: "riido-control-plane-api-client-delivery-evidence.v1",
		ID:            m.ID,
		Status:        "verified",
		Result:        r,
		Workflow:      m.Workflow,
		GeneratedDoc:  m.GeneratedDoc,
		Loop:          m.Loop,
	}
}
