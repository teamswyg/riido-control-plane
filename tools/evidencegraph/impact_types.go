package main

type impactEvidence struct {
	Enabled           bool          `json:"enabled"`
	BaseRef           string        `json:"base_ref,omitempty"`
	ChangedFileCount  int           `json:"changed_file_count"`
	ChangedFiles      []string      `json:"changed_files,omitempty"`
	AddedChainCount   int           `json:"added_chain_count"`
	ChangedChainCount int           `json:"changed_chain_count"`
	RemovedChainCount int           `json:"removed_chain_count"`
	AddedChains       []impactChain `json:"added_chains,omitempty"`
	ChangedChains     []impactChain `json:"changed_chains,omitempty"`
	RemovedChains     []impactChain `json:"removed_chains,omitempty"`
}

type impactChain struct {
	ID                    string   `json:"id"`
	ChangedExecutableRefs []string `json:"changed_executable_refs"`
	Claims                []string `json:"claims,omitempty"`
	VerifierRefs          []string `json:"verifier_refs,omitempty"`
	VerifierCommands      []string `json:"verifier_commands,omitempty"`
	EvidenceRefs          []string `json:"evidence_refs,omitempty"`
	NextLoop              string   `json:"next_loop,omitempty"`
}
