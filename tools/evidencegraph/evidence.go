package main

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	ChainCount       int          `json:"chain_count"`
	ClaimRefs        int          `json:"claim_ref_count"`
	ChangeRefs       int          `json:"change_ref_count"`
	VerifierRefs     int          `json:"verifier_ref_count"`
	EvidenceRefs     int          `json:"evidence_ref_count"`
	Workflow         string       `json:"workflow"`
	GeneratedDoc     string       `json:"generated_doc"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	LoopRegistry     string       `json:"loop_registry_manifest"`
	Loop             loopRecord   `json:"loop"`
	Chains           []chainBrief `json:"chains"`
}

type chainBrief struct {
	ID       string `json:"id"`
	NextLoop string `json:"next_loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		ChainCount:       result.Chains,
		ClaimRefs:        result.ClaimRefs,
		ChangeRefs:       result.ChangeRefs,
		VerifierRefs:     result.VerifierRefs,
		EvidenceRefs:     result.EvidenceRefs,
		Workflow:         m.Workflow,
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.Evidence,
		LoopRegistry:     m.LoopRegistry,
		Loop:             m.Loop,
		Chains:           briefChains(m.Chains),
	}
}

func briefChains(chains []chain) []chainBrief {
	out := make([]chainBrief, 0, len(chains))
	for _, c := range chains {
		out = append(out, chainBrief{ID: c.ID, NextLoop: c.NextLoop})
	}
	return out
}
