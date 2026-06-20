package main

import "encoding/json"

type sourceManifest struct {
	SchemaVersion                   string            `json:"schema_version"`
	ID                              string            `json:"id"`
	RiidoTask                       string            `json:"riido_task"`
	StabilizedBy                    []string          `json:"stabilized_by"`
	HumanDoc                        string            `json:"human_doc"`
	RelatedManifests                []string          `json:"related_manifests"`
	Figma                           json.RawMessage   `json:"figma"`
	InspectionMethod                json.RawMessage   `json:"inspection_method"`
	SupportingToolLimitations       []json.RawMessage `json:"supporting_tool_limitations"`
	CoveragePolicy                  json.RawMessage   `json:"coverage_policy"`
	AnnotationPolicy                annotationPolicy  `json:"api_generated_annotation_content_policy"`
	ExpectedPages                   []json.RawMessage `json:"expected_pages"`
	ExpectedTopLevelNodes           []json.RawMessage `json:"expected_top_level_nodes"`
	NonUITopLevelInventory          []json.RawMessage `json:"non_ui_top_level_inventory"`
	VerifiedEvidenceNodes           []json.RawMessage `json:"verified_evidence_nodes"`
	NonUITopLevelNodes              []json.RawMessage `json:"non_ui_top_level_nodes"`
	APIGeneratedAnnotations         []json.RawMessage `json:"api_generated_annotations"`
	APIGeneratedAnnotationInventory []json.RawMessage `json:"api_generated_annotation_inventory"`
	Entries                         []json.RawMessage `json:"entries"`
}

type annotationPolicy struct {
	CategoryID        string            `json:"category_id"`
	CategoryLabel     string            `json:"category_label"`
	LabelFormat       []string          `json:"label_format"`
	Rule              string            `json:"rule"`
	RetiredCategories []json.RawMessage `json:"retired_categories"`
	LiveInspection    liveInspection    `json:"live_inspection"`
}

type liveInspection struct {
	ObservedAt                   string            `json:"observed_at"`
	Tool                         string            `json:"tool"`
	PageCounts                   []json.RawMessage `json:"page_counts"`
	TotalRiidoAnnotations        int               `json:"total_riido_annotations"`
	TotalAPIGeneratedAnnotations int               `json:"total_api_generated_annotations"`
}
