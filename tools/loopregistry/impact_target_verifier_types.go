package main

type targetVerifierPlan struct {
	ChangedPathCount int                       `json:"changed_path_count"`
	MatchedPathCount int                       `json:"matched_path_count"`
	ComponentCount   int                       `json:"component_count"`
	CommandCount     int                       `json:"command_count"`
	VerifierCommands []string                  `json:"verifier_commands"`
	Components       []targetVerifierComponent `json:"components"`
	Paths            []targetVerifierPath      `json:"paths"`
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
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}
