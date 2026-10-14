package main

type figmaProjectionManifest struct {
	SchemaVersion string `json:"schema_version"`
	ID            string `json:"id"`
	RiidoTask     string `json:"riido_task"`
	GeneratedReaderMetadata
	EvidenceTool                      string                                    `json:"evidence_tool"`
	SourceContractsManifest           figmaProjectionSourceManifest             `json:"source_contracts_manifest"`
	ProjectionPolicy                  figmaProjectionPolicy                     `json:"projection_policy"`
	MirroredSupportingToolLimitations []figmaProjectionSupportingToolLimitation `json:"mirrored_supporting_tool_limitations"`
	LegacyNonUIAbsorptions            []figmaProjectionLegacyAbsorption         `json:"legacy_non_ui_absorptions"`
	NonUIPlanningAbsorptions          []figmaProjectionPlanningAbsorption       `json:"non_ui_planning_absorptions"`
	Entries                           []figmaProjectionEntry                    `json:"entries"`
	Loop                              figmaProjectionLoop                       `json:"loop"`
}

type figmaProjectionLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type figmaProjectionSourceManifest struct {
	Repo          string   `json:"repo"`
	Path          string   `json:"path"`
	SchemaVersion string   `json:"schema_version"`
	ID            string   `json:"id"`
	StabilizedBy  []string `json:"stabilized_by"`
}

type figmaProjectionPolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}

type figmaProjectionSupportingToolLimitation struct {
	SourceID                     string   `json:"source_id"`
	LocalScope                   string   `json:"local_scope"`
	RequiredAuthoritativePages   []string `json:"required_authoritative_pages,omitempty"`
	RequiredAuthoritativeResults []string `json:"required_authoritative_results,omitempty"`
	ForbiddenProjectionEffects   []string `json:"forbidden_projection_effects"`
}
