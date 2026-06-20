package main

type figmaSourceCoverageEntry struct {
	PageID                   string                       `json:"page_id,omitempty"`
	NodeID                   string                       `json:"node_id"`
	Name                     string                       `json:"name,omitempty"`
	CoverageStatus           string                       `json:"coverage_status,omitempty"`
	EvidenceKind             string                       `json:"evidence_kind,omitempty"`
	AbsorbedByTopLevelNodeID string                       `json:"absorbed_by_top_level_node_id,omitempty"`
	SSOTDocs                 []string                     `json:"ssot_docs,omitempty"`
	OwnerRepos               []string                     `json:"owner_repos,omitempty"`
	GeneratedPaths           []string                     `json:"generated_paths,omitempty"`
	CoveredFacts             []string                     `json:"covered_facts,omitempty"`
	DirectionLoop            figmaSourceCoverageDirection `json:"direction_loop,omitempty"`
	Reason                   string                       `json:"reason,omitempty"`
}

type figmaSourceCoverageDirection struct {
	TopDown  string `json:"top_down,omitempty"`
	BottomUp string `json:"bottom_up,omitempty"`
}
