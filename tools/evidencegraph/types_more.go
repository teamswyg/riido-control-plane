package main

type ref struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Redacted bool   `json:"redacted,omitempty"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type verifyResult struct {
	Chains       int `json:"chains"`
	ClaimRefs    int `json:"claim_refs"`
	ChangeRefs   int `json:"change_refs"`
	VerifierRefs int `json:"verifier_refs"`
	EvidenceRefs int `json:"evidence_refs"`
}
