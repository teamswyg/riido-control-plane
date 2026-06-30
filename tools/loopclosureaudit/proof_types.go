package main

type proof struct {
	Kind    string        `json:"kind"`
	Key     string        `json:"key"`
	Status  string        `json:"status"`
	Surface *proofSurface `json:"surface,omitempty"`
}

type proofSurface struct {
	Files                 []string `json:"files,omitempty"`
	Verifiers             []string `json:"verifiers,omitempty"`
	GeneratedDocs         []string `json:"generated_docs,omitempty"`
	SemanticHash          string   `json:"semantic_hash,omitempty"`
	Claims                []string `json:"claims,omitempty"`
	Workflow              string   `json:"workflow,omitempty"`
	Contains              []string `json:"contains,omitempty"`
	Observes              []string `json:"observes,omitempty"`
	Verifies              []string `json:"verifies,omitempty"`
	FailsWhen             []string `json:"fails_when,omitempty"`
	RefreshWorkflow       string   `json:"refresh_workflow,omitempty"`
	ExpiresAfterHours     int      `json:"expires_after_hours,omitempty"`
	Providers             []string `json:"providers,omitempty"`
	PromotesTo            []string `json:"promotes_to,omitempty"`
	From                  string   `json:"from,omitempty"`
	To                    string   `json:"to,omitempty"`
	Relation              string   `json:"relation,omitempty"`
	PreCommitHook         string   `json:"pre_commit_hook,omitempty"`
	Summary               string   `json:"summary,omitempty"`
	GraphChainCount       int      `json:"graph_chain_count,omitempty"`
	GraphCompleteChains   int      `json:"graph_complete_chain_count,omitempty"`
	GraphClaimBoundChains int      `json:"graph_claim_bound_chain_count,omitempty"`
	GraphUnclaimedChains  int      `json:"graph_unclaimed_chain_count,omitempty"`
	GraphNextLoopCount    int      `json:"graph_next_loop_count,omitempty"`
	GraphNextLoops        []string `json:"graph_next_loops,omitempty"`
}
