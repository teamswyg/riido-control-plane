package main

type evidence struct {
	SchemaVersion   string     `json:"schema_version"`
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	PreCommitConfig string     `json:"pre_commit_config"`
	Workflow        string     `json:"workflow"`
	GeneratedDoc    string     `json:"generated_doc"`
	HookCount       int        `json:"hook_count"`
	ScriptCount     int        `json:"script_count"`
	PhraseChecks    int        `json:"phrase_checks"`
	Loop            loopRecord `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:   evidenceSchema,
		ID:              m.ID,
		Status:          "verified",
		PreCommitConfig: m.PreCommitConfig,
		Workflow:        m.Workflow,
		GeneratedDoc:    m.GeneratedDoc,
		HookCount:       result.Hooks,
		ScriptCount:     result.Scripts,
		PhraseChecks:    result.PhraseChecks,
		Loop:            m.Loop,
	}
}
