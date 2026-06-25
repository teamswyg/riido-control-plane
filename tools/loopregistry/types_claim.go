package main

type claimBinding struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Loop           string   `json:"loop"`
	CoversObserves []string `json:"covers_observes,omitempty"`
	CoversVerifies []string `json:"covers_verifies,omitempty"`
	CoversFails    []string `json:"covers_fails_when,omitempty"`
	CoversEvidence []string `json:"covers_evidence,omitempty"`
	Files          []string `json:"files"`
	Verifiers      []string `json:"verifiers"`
	GeneratedDoc   []string `json:"generated_docs"`
	SemanticHash   string   `json:"semantic_hash"`
}

type claimSurface struct {
	ID               string   `json:"id"`
	CodePaths        []string `json:"code_paths"`
	TestPaths        []string `json:"test_paths"`
	ManifestPaths    []string `json:"manifest_paths"`
	GeneratedDocs    []string `json:"generated_docs"`
	CoversObserves   []string `json:"covers_observes"`
	CoversVerifies   []string `json:"covers_verifies"`
	CoversFails      []string `json:"covers_fails_when"`
	CoversEvidence   []string `json:"covers_evidence"`
	Verifiers        []string `json:"verifiers"`
	VerifierCommands []string `json:"verifier_commands"`
	EvidenceChainIDs []string `json:"evidence_chain_ids"`
}
