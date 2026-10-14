package main

type figmaProjectionEntry struct {
	NodeID                          string   `json:"node_id"`
	Name                            string   `json:"name"`
	ProjectionStatus                string   `json:"projection_status"`
	SourceCoverageStatus            string   `json:"source_coverage_status"`
	LocalScope                      string   `json:"local_scope,omitempty"`
	RequiredGeneratedPaths          []string `json:"required_generated_paths,omitempty"`
	ForbiddenGeneratedPathFragments []string `json:"forbidden_generated_path_fragments,omitempty"`
	NoEndpointReason                string   `json:"no_endpoint_reason,omitempty"`
}

type figmaProjectionLegacyAbsorption struct {
	NodeID                   string   `json:"node_id"`
	Name                     string   `json:"name"`
	ProjectionStatus         string   `json:"projection_status"`
	SourceCoverageStatus     string   `json:"source_coverage_status"`
	AbsorbedByTopLevelNodeID string   `json:"absorbed_by_top_level_node_id"`
	LocalScope               string   `json:"local_scope"`
	RequiredGeneratedPaths   []string `json:"required_generated_paths"`
}

type figmaProjectionPlanningAbsorption struct {
	NodeID                 string   `json:"node_id"`
	Name                   string   `json:"name"`
	ProjectionStatus       string   `json:"projection_status"`
	SourceCoverageStatus   string   `json:"source_coverage_status"`
	LocalScope             string   `json:"local_scope"`
	RequiredGeneratedPaths []string `json:"required_generated_paths"`
	NoNewEndpointReason    string   `json:"no_new_endpoint_reason"`
}
