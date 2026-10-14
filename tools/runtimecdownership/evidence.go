package main

type evidence struct {
	SchemaVersion   string       `json:"schema_version"`
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	Manifest        string       `json:"manifest"`
	GeneratedDoc    string       `json:"generated_doc"`
	Strategies      int          `json:"strategies"`
	PublicPolicies  int          `json:"public_policies"`
	PublicGuards    int          `json:"public_guards"`
	ForbiddenItems  int          `json:"forbidden_items"`
	InfraLinks      int          `json:"infra_links"`
	LoopFields      int          `json:"loop_fields"`
	StableKeyCount  int          `json:"stable_key_count"`
	WorkflowCount   int          `json:"workflow_count"`
	HardeningCount  int          `json:"hardening_count"`
	SupersedesCount int          `json:"supersedes_count"`
	Loop            evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion: evidenceSchema, ID: m.ID, Status: "verified",
		Manifest: defaultManifest, GeneratedDoc: generatedDoc,
		Strategies: result.Strategies, PublicPolicies: result.PublicPolicies,
		PublicGuards: result.PublicGuards, ForbiddenItems: result.ForbiddenItems,
		InfraLinks: result.InfraLinks, LoopFields: result.LoopFields,
		StableKeyCount: result.StableKeyCount, WorkflowCount: result.WorkflowCount,
		HardeningCount: result.HardeningCount, SupersedesCount: result.SupersedesCount,
		Loop: m.Loop,
	}
}
