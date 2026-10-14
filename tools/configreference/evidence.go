package main

type evidence struct {
	SchemaVersion   string       `json:"schema_version"`
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	RuntimeEnvCount int          `json:"runtime_env_count"`
	ManifestCount   int          `json:"manifest_count"`
	SecretCount     int          `json:"secret_count"`
	OperatorCount   int          `json:"operator_count"`
	Workflow        string       `json:"workflow"`
	GeneratedDoc    string       `json:"generated_doc"`
	Loop            loopEvidence `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:   evidenceSchema,
		ID:              m.ID,
		Status:          "verified",
		RuntimeEnvCount: result.RuntimeEnvCount,
		ManifestCount:   result.ManifestCount,
		SecretCount:     result.SecretCount,
		OperatorCount:   result.OperatorCount,
		Workflow:        m.Workflow,
		GeneratedDoc:    m.GeneratedDoc,
		Loop:            m.Loop,
	}
}
