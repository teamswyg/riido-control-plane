package main

type targetVerifierPlan struct {
	ChangedPathCount int                  `json:"changed_path_count"`
	MatchedPathCount int                  `json:"matched_path_count"`
	CommandCount     int                  `json:"command_count"`
	VerifierCommands []string             `json:"verifier_commands"`
	Paths            []targetVerifierPath `json:"paths"`
}

type targetVerifierPath struct {
	Path             string   `json:"path"`
	Kind             string   `json:"kind"`
	LoopIDs          []string `json:"loop_ids"`
	ClaimIDs         []string `json:"claim_ids"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}
