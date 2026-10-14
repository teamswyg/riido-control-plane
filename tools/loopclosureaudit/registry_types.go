package main

type loopRegistry struct {
	Loops  []registryLoop  `json:"loops"`
	Claims []registryClaim `json:"claim_bindings"`
	Edges  []graphEdge     `json:"evidence_graph"`
}

type registryLoop struct {
	ID                string                `json:"id"`
	Kind              string                `json:"kind"`
	Observes          []string              `json:"observes"`
	Verifies          []string              `json:"verifies"`
	Evidence          []registryEvidenceRef `json:"evidence"`
	RefreshWorkflow   string                `json:"refresh_workflow"`
	FailsWhen         []string              `json:"fails_when"`
	Providers         []string              `json:"providers"`
	PromotesTo        []string              `json:"promotes_to"`
	ExpiresAfterHours int                   `json:"expires_after_hours"`
}

type registryEvidenceRef struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Redacted bool   `json:"redacted"`
}

type registryClaim struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Loop           string   `json:"loop"`
	CoversObserves []string `json:"covers_observes,omitempty"`
	CoversVerifies []string `json:"covers_verifies,omitempty"`
	CoversFails    []string `json:"covers_fails_when,omitempty"`
	Files          []string `json:"files"`
	Verifiers      []string `json:"verifiers"`
	GeneratedDocs  []string `json:"generated_docs"`
	SemanticHash   string   `json:"semantic_hash"`
}

type graphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type graphChainSummary struct {
	ChainCount       int `json:"chain_count"`
	CompleteChains   int `json:"complete_chain_count"`
	ClaimBoundChains int `json:"claim_bound_chain_count"`
	UnclaimedChains  int `json:"unclaimed_chain_count"`
	NextLoopCount    int `json:"next_loop_count"`
}

type graphNextLoopSummary struct {
	NextLoop   string `json:"next_loop"`
	ChainCount int    `json:"chain_count"`
}

type preCommitManifest struct {
	Hooks []preCommitHook `json:"hooks"`
}

type preCommitHook struct {
	ID       string   `json:"id"`
	Summary  string   `json:"summary"`
	Contains []string `json:"contains"`
}
