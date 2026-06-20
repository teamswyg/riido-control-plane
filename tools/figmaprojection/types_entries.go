package main

type sourceEntry struct {
	PageID                   string            `json:"page_id,omitempty"`
	NodeID                   string            `json:"node_id"`
	Name                     string            `json:"name,omitempty"`
	CoverageStatus           string            `json:"coverage_status,omitempty"`
	EvidenceKind             string            `json:"evidence_kind,omitempty"`
	AbsorbedByTopLevelNodeID string            `json:"absorbed_by_top_level_node_id,omitempty"`
	SSOTDocs                 []string          `json:"ssot_docs,omitempty"`
	OwnerRepos               []string          `json:"owner_repos,omitempty"`
	GeneratedPaths           []string          `json:"generated_paths,omitempty"`
	CoveredFacts             []string          `json:"covered_facts,omitempty"`
	DirectionLoop            map[string]string `json:"direction_loop,omitempty"`
	Reason                   string            `json:"reason,omitempty"`
}

type apiAnnotation struct {
	NodeID                 string `json:"node_id"`
	TopLevelNodeID         string `json:"top_level_node_id"`
	CoverageEntryNodeID    string `json:"coverage_entry_node_id"`
	CategoryID             string `json:"category_id"`
	CategoryLabel          string `json:"category_label"`
	FigmaLabel             string `json:"figma_label"`
	FigmaGeneratedPath     string `json:"figma_generated_path"`
	CanonicalGeneratedPath string `json:"canonical_generated_path"`
	ResolutionStatus       string `json:"resolution_status"`
	Resolution             string `json:"resolution"`
}
