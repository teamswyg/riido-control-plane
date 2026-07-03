package main

type projectionManifest struct {
	SchemaVersion       string                 `json:"schema_version"`
	ID                  string                 `json:"id"`
	RiidoTask           string                 `json:"riido_task"`
	GeneratedDoc        string                 `json:"generated_doc"`
	Workflow            string                 `json:"workflow"`
	EvidenceArtifact    string                 `json:"evidence_artifact"`
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

type evidence struct {
	SchemaVersion                string       `json:"schema_version"`
	ID                           string       `json:"id"`
	Status                       string       `json:"status"`
	ProjectionEntries            int          `json:"projection_entries"`
	LegacyAbsorptions            int          `json:"legacy_absorptions"`
	PlanningAbsorptions          int          `json:"planning_absorptions"`
	MirroredLimitations          int          `json:"mirrored_limitations"`
	SourceExpectedPages          int          `json:"source_expected_pages"`
	SourceNonUITopLevelNodes     int          `json:"source_non_ui_top_level_nodes"`
	SourceGeneratedInventory     int          `json:"source_generated_inventory"`
	TotalRiidoAnnotations        int          `json:"total_riido_annotations"`
	TotalAPIGeneratedAnnotations int          `json:"total_api_generated_annotations"`
	SourceStabilizedBy           int          `json:"source_stabilized_by"`
	Loop                         evidenceLoop `json:"loop"`
}

func newEvidence(p projectionManifest, s sourceManifest) evidence {
	return evidence{
		SchemaVersion: evidenceSchema, ID: p.ID, Status: "verified",
		ProjectionEntries: len(p.Entries), LegacyAbsorptions: len(p.LegacyAbsorptions),
		PlanningAbsorptions: len(p.PlanningAbsorptions), MirroredLimitations: len(p.ToolLimitations),
		SourceExpectedPages: len(s.ExpectedPages), SourceNonUITopLevelNodes: len(s.NonUITopLevelNodes),
		SourceGeneratedInventory:     len(s.APIGeneratedAnnotationInventory),
		TotalRiidoAnnotations:        s.AnnotationPolicy.LiveInspection.TotalRiidoAnnotations,
		TotalAPIGeneratedAnnotations: s.AnnotationPolicy.LiveInspection.TotalAPIGeneratedAnnotations,
		SourceStabilizedBy:           len(s.StabilizedBy), Loop: p.Loop,
	}
}
