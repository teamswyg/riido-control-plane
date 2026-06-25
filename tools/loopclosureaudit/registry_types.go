package main

type loopRegistry struct {
	Loops  []registryLoop  `json:"loops"`
	Claims []registryClaim `json:"claim_bindings"`
	Edges  []graphEdge     `json:"evidence_graph"`
}

type registryLoop struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	RefreshWorkflow   string   `json:"refresh_workflow"`
	Providers         []string `json:"providers"`
	PromotesTo        []string `json:"promotes_to"`
	ExpiresAfterHours int      `json:"expires_after_hours"`
}

type registryClaim struct {
	ID            string   `json:"id"`
	Statement     string   `json:"statement"`
	Loop          string   `json:"loop"`
	Files         []string `json:"files"`
	Verifiers     []string `json:"verifiers"`
	GeneratedDocs []string `json:"generated_docs"`
	SemanticHash  string   `json:"semantic_hash"`
}

type graphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type evidenceGraph struct {
	Chains []evidenceChain `json:"chains"`
}

type evidenceChain struct {
	ID     string   `json:"id"`
	Claims []string `json:"claims"`
}

type preCommitManifest struct {
	Hooks []preCommitHook `json:"hooks"`
}

type preCommitHook struct {
	ID string `json:"id"`
}
