package main

type candidateSubject struct {
	Kind              string   `json:"kind"`
	LoopID            string   `json:"loop_id,omitempty"`
	CommandKind       string   `json:"command_kind,omitempty"`
	Command           string   `json:"command,omitempty"`
	SourceCandidateID string   `json:"source_candidate_id,omitempty"`
	SourceSubjectKind string   `json:"source_subject_kind,omitempty"`
	SourcePath        string   `json:"source_path,omitempty"`
	GeneratedAt       string   `json:"generated_at,omitempty"`
	ExpiresAt         string   `json:"expires_at,omitempty"`
	Reason            string   `json:"reason,omitempty"`
	ClaimIDs          []string `json:"claim_ids,omitempty"`
	EvidenceChainIDs  []string `json:"evidence_chain_ids,omitempty"`
}
