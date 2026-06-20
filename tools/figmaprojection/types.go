package main

import "encoding/json"

type projectionManifest struct {
	SchemaVersion       string            `json:"schema_version"`
	ID                  string            `json:"id"`
	RiidoTask           string            `json:"riido_task"`
	EvidenceTool        string            `json:"evidence_tool"`
	Source              sourcePointer     `json:"source_contracts_manifest"`
	ProjectionPolicy    json.RawMessage   `json:"projection_policy"`
	ToolLimitations     []json.RawMessage `json:"mirrored_supporting_tool_limitations"`
	LegacyAbsorptions   []json.RawMessage `json:"legacy_non_ui_absorptions"`
	PlanningAbsorptions []json.RawMessage `json:"non_ui_planning_absorptions"`
	Entries             []projectionEntry `json:"entries"`
	Loop                evidenceLoop      `json:"loop"`
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
