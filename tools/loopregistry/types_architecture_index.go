package main

type architectureIndex struct {
	PathCount            int                       `json:"path_count"`
	ComponentCount       int                       `json:"component_count"`
	BindingCount         int                       `json:"binding_count"`
	VerifierCommandCount int                       `json:"verifier_command_count"`
	Components           []architectureComponent   `json:"components"`
	Paths                []architecturePathBinding `json:"paths"`
}

type architectureComponent struct {
	Component        string   `json:"component"`
	PathCount        int      `json:"path_count"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}

type architecturePathBinding struct {
	Path             string   `json:"path"`
	Kind             string   `json:"kind"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}
