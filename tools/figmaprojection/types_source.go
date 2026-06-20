package main

type sourceManifest struct {
	SchemaVersion                   string           `json:"schema_version"`
	ID                              string           `json:"id"`
	RiidoTask                       string           `json:"riido_task"`
	StabilizedBy                    []string         `json:"stabilized_by"`
	HumanDoc                        string           `json:"human_doc"`
	RelatedManifests                []string         `json:"related_manifests"`
	Figma                           sourceFigma      `json:"figma"`
	InspectionMethod                inspectionMethod `json:"inspection_method"`
	SupportingToolLimitations       []toolLimitation `json:"supporting_tool_limitations"`
	CoveragePolicy                  projectionPolicy `json:"coverage_policy"`
	AnnotationPolicy                annotationPolicy `json:"api_generated_annotation_content_policy"`
	ExpectedPages                   []coverageNode   `json:"expected_pages"`
	ExpectedTopLevelNodes           []coverageNode   `json:"expected_top_level_nodes"`
	NonUITopLevelInventory          []inventoryPage  `json:"non_ui_top_level_inventory"`
	VerifiedEvidenceNodes           []coverageNode   `json:"verified_evidence_nodes"`
	NonUITopLevelNodes              []sourceEntry    `json:"non_ui_top_level_nodes"`
	APIGeneratedAnnotations         []apiAnnotation  `json:"api_generated_annotations"`
	APIGeneratedAnnotationInventory []apiGroup       `json:"api_generated_annotation_inventory"`
	Entries                         []sourceEntry    `json:"entries"`
}

type annotationPolicy struct {
	CategoryID        string            `json:"category_id"`
	CategoryLabel     string            `json:"category_label"`
	LabelFormat       []string          `json:"label_format"`
	Rule              string            `json:"rule"`
	RetiredCategories []retiredCategory `json:"retired_categories"`
	LiveInspection    liveInspection    `json:"live_inspection"`
}

type liveInspection struct {
	ObservedAt                   string      `json:"observed_at"`
	Tool                         string      `json:"tool"`
	PageCounts                   []pageCount `json:"page_counts"`
	TotalRiidoAnnotations        int         `json:"total_riido_annotations"`
	TotalAPIGeneratedAnnotations int         `json:"total_api_generated_annotations"`
}
