package main

type apiGroup struct {
	UIArea                 string      `json:"ui_area"`
	CategoryID             string      `json:"category_id"`
	CategoryLabel          string      `json:"category_label"`
	FigmaGeneratedPath     string      `json:"figma_generated_path"`
	CanonicalGeneratedPath string      `json:"canonical_generated_path"`
	OperationKind          string      `json:"operation_kind"`
	Background             string      `json:"background"`
	AnnotationCount        int         `json:"annotation_count"`
	Sources                []apiSource `json:"sources"`
}

type apiSource struct {
	PageID              string   `json:"page_id"`
	TopLevelNodeID      string   `json:"top_level_node_id"`
	CoverageEntryNodeID string   `json:"coverage_entry_node_id"`
	NodeIDs             []string `json:"node_ids"`
}

type pageCount struct {
	PageID               string `json:"page_id"`
	PageName             string `json:"page_name"`
	RiidoAnnotationCount int    `json:"riido_annotation_count"`
	APIGeneratedCount    int    `json:"api_generated_count"`
	MissingOperationKind int    `json:"missing_operation_kind"`
	MissingBackground    int    `json:"missing_background"`
}

type retiredCategory struct {
	CategoryID       string `json:"category_id"`
	CategoryLabel    string `json:"category_label"`
	RetirementStatus string `json:"retirement_status"`
	LiveUsageCount   int    `json:"live_usage_count"`
	ObservedAt       string `json:"observed_at"`
	ToolLimitation   string `json:"tool_limitation"`
}
