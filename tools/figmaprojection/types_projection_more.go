package main

type projectionLimitation struct {
	SourceID                     string   `json:"source_id"`
	LocalScope                   string   `json:"local_scope"`
	RequiredAuthoritativePages   []string `json:"required_authoritative_pages,omitempty"`
	RequiredAuthoritativeResults []string `json:"required_authoritative_results,omitempty"`
	ForbiddenProjectionEffects   []string `json:"forbidden_projection_effects"`
}

type legacyAbsorption struct {
	NodeID                   string   `json:"node_id"`
	Name                     string   `json:"name"`
	ProjectionStatus         string   `json:"projection_status"`
	SourceCoverageStatus     string   `json:"source_coverage_status"`
	AbsorbedByTopLevelNodeID string   `json:"absorbed_by_top_level_node_id"`
	LocalScope               string   `json:"local_scope"`
	RequiredGeneratedPaths   []string `json:"required_generated_paths"`
}

type planningAbsorption struct {
	NodeID                 string   `json:"node_id"`
	Name                   string   `json:"name"`
	ProjectionStatus       string   `json:"projection_status"`
	SourceCoverageStatus   string   `json:"source_coverage_status"`
	LocalScope             string   `json:"local_scope"`
	RequiredGeneratedPaths []string `json:"required_generated_paths"`
	NoNewEndpointReason    string   `json:"no_new_endpoint_reason"`
}
