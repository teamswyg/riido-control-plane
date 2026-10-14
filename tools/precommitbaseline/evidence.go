package main

type evidence struct {
	SchemaVersion     string     `json:"schema_version"`
	ID                string     `json:"id"`
	Status            string     `json:"status"`
	GeneratedAt       string     `json:"generated_at"`
	ExpiresAt         string     `json:"expires_at"`
	EvidenceTTL       int        `json:"evidence_ttl_hours"`
	PreCommitConfig   string     `json:"pre_commit_config"`
	Workflow          string     `json:"workflow"`
	GeneratedDoc      string     `json:"generated_doc"`
	HookCount         int        `json:"hook_count"`
	ScriptCount       int        `json:"script_count"`
	PhraseChecks      int        `json:"phrase_checks"`
	WorkflowScheduled bool       `json:"workflow_scheduled"`
	Loop              loopRecord `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	generatedAt, expiresAt := evidenceWindow(m.EvidenceTTL)
	return evidence{
		SchemaVersion:     evidenceSchema,
		ID:                m.ID,
		Status:            "verified",
		GeneratedAt:       generatedAt,
		ExpiresAt:         expiresAt,
		EvidenceTTL:       m.EvidenceTTL,
		PreCommitConfig:   m.PreCommitConfig,
		Workflow:          m.Workflow,
		GeneratedDoc:      m.GeneratedDoc,
		HookCount:         result.Hooks,
		ScriptCount:       result.Scripts,
		PhraseChecks:      result.PhraseChecks,
		WorkflowScheduled: result.WorkflowScheduled,
		Loop:              m.Loop,
	}
}
