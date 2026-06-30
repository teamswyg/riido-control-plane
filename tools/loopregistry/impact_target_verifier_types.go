package main

type targetVerifierPlan struct {
	ChangedPathCount     int                       `json:"changed_path_count"`
	MatchedPathCount     int                       `json:"matched_path_count"`
	ExactPathCount       int                       `json:"exact_path_count"`
	ComponentRouteCount  int                       `json:"component_route_count"`
	ComponentCount       int                       `json:"component_count"`
	CommandCount         int                       `json:"command_count"`
	RunnableCommandCount int                       `json:"runnable_command_count"`
	FocusedClaimIDs      []string                  `json:"focused_claim_ids,omitempty"`
	FocusedCommandCount  int                       `json:"focused_command_count,omitempty"`
	FocusedCommands      []string                  `json:"focused_commands,omitempty"`
	RunnableCommands     []string                  `json:"runnable_commands"`
	EntrypointCommands   []string                  `json:"entrypoint_commands"`
	VerifierCommands     []string                  `json:"verifier_commands"`
	CommandUnits         []targetVerifierCommand   `json:"command_units"`
	Components           []targetVerifierComponent `json:"components"`
	Paths                []targetVerifierPath      `json:"paths"`
}

type targetVerifierCommand struct {
	Command          string   `json:"command"`
	PathCount        int      `json:"path_count"`
	ComponentCount   int      `json:"component_count"`
	Paths            []string `json:"paths"`
	Components       []string `json:"components"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}

type targetVerifierComponent struct {
	Component        string   `json:"component"`
	PathCount        int      `json:"path_count"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}

type targetVerifierPath struct {
	Path             string   `json:"path"`
	Component        string   `json:"component"`
	Kind             string   `json:"kind"`
	MatchKind        string   `json:"match_kind"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}
