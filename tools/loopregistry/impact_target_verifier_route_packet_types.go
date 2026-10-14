package main

type targetVerifierRoute struct {
	Path                   string   `json:"path"`
	Component              string   `json:"component"`
	MatchKind              string   `json:"match_kind"`
	RunnableCommands       []string `json:"runnable_commands"`
	RunnableCommandCount   int      `json:"runnable_command_count"`
	DirectCommandCount     int      `json:"direct_command_count"`
	UsesEntrypointFallback bool     `json:"uses_entrypoint_fallback"`
	LoopIDs                []string `json:"loop_ids"`
	ClaimIDs               []string `json:"claim_ids"`
	EvidenceChainIDs       []string `json:"evidence_chain_ids"`
}
