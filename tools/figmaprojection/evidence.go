package main

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
