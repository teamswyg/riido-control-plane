package main

type projectionManifest struct {
	SchemaVersion       string                 `json:"schema_version"`
	ID                  string                 `json:"id"`
	RiidoTask           string                 `json:"riido_task"`
	EvidenceTool        string                 `json:"evidence_tool"`
	Source              sourcePointer          `json:"source_contracts_manifest"`
	ProjectionPolicy    projectionPolicy       `json:"projection_policy"`
	ToolLimitations     []projectionLimitation `json:"mirrored_supporting_tool_limitations"`
	LegacyAbsorptions   []legacyAbsorption     `json:"legacy_non_ui_absorptions"`
	PlanningAbsorptions []planningAbsorption   `json:"non_ui_planning_absorptions"`
	Entries             []projectionEntry      `json:"entries"`
	Loop                evidenceLoop           `json:"loop"`
}

type sourcePointer struct {
	Repo          string   `json:"repo"`
	Path          string   `json:"path"`
	SchemaVersion string   `json:"schema_version"`
	ID            string   `json:"id"`
	StabilizedBy  []string `json:"stabilized_by"`
}

type projectionEntry struct {
	NodeID                          string   `json:"node_id"`
	Name                            string   `json:"name"`
	ProjectionStatus                string   `json:"projection_status"`
	SourceCoverageStatus            string   `json:"source_coverage_status"`
	LocalScope                      string   `json:"local_scope,omitempty"`
	RequiredGeneratedPaths          []string `json:"required_generated_paths,omitempty"`
	ForbiddenGeneratedPathFragments []string `json:"forbidden_generated_path_fragments,omitempty"`
	NoEndpointReason                string   `json:"no_endpoint_reason,omitempty"`
}

type projectionPolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}
