package main

type figmaSourceCoverageManifest struct {
	SchemaVersion                       string                                       `json:"schema_version"`
	ID                                  string                                       `json:"id"`
	RiidoTask                           string                                       `json:"riido_task"`
	StabilizedBy                        []string                                     `json:"stabilized_by"`
	HumanDoc                            string                                       `json:"human_doc"`
	RelatedManifests                    []string                                     `json:"related_manifests"`
	Figma                               figmaSourceCoverageSource                    `json:"figma"`
	InspectionMethod                    figmaCoverageInspectionMethod                `json:"inspection_method"`
	SupportingToolLimitations           []figmaSourceSupportingToolLimitation        `json:"supporting_tool_limitations"`
	CoveragePolicy                      figmaSourceCoveragePolicy                    `json:"coverage_policy"`
	APIGeneratedAnnotationContentPolicy figmaSourceAPIGeneratedAnnotationContentRule `json:"api_generated_annotation_content_policy"`
	ExpectedPages                       []figmaSourceCoveragePage                    `json:"expected_pages"`
	ExpectedTopLevelNodes               []figmaSourceCoverageNode                    `json:"expected_top_level_nodes"`
	NonUITopLevelInventory              []figmaSourceCoverageInventory               `json:"non_ui_top_level_inventory"`
	VerifiedEvidenceNodes               []figmaSourceCoverageNode                    `json:"verified_evidence_nodes"`
	NonUITopLevelNodes                  []figmaSourceCoverageEntry                   `json:"non_ui_top_level_nodes"`
	APIGeneratedAnnotations             []figmaSourceAPIGeneratedAnnotation          `json:"api_generated_annotations"`
	APIGeneratedAnnotationInventory     []figmaSourceAPIGeneratedAnnotationGroup     `json:"api_generated_annotation_inventory"`
	Entries                             []figmaSourceCoverageEntry                   `json:"entries"`
}

type figmaSourceCoverageSource struct {
	FileKey          string `json:"file_key"`
	FileName         string `json:"file_name"`
	PageID           string `json:"page_id"`
	PageName         string `json:"page_name"`
	InspectedAt      string `json:"inspected_at"`
	InspectionSource string `json:"inspection_source"`
}

type figmaSourceCoveragePolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}
