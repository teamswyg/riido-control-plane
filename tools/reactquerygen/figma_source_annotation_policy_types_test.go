package main

type figmaSourceAPIGeneratedAnnotationContentRule struct {
	CategoryID        string                                             `json:"category_id"`
	CategoryLabel     string                                             `json:"category_label"`
	LabelFormat       []string                                           `json:"label_format"`
	Rule              string                                             `json:"rule"`
	RetiredCategories []figmaSourceAPIGeneratedAnnotationRetiredCategory `json:"retired_categories"`
	LiveInspection    figmaSourceAPIGeneratedAnnotationLiveScan          `json:"live_inspection"`
}

type figmaSourceAPIGeneratedAnnotationRetiredCategory struct {
	CategoryID       string `json:"category_id"`
	CategoryLabel    string `json:"category_label"`
	RetirementStatus string `json:"retirement_status"`
	LiveUsageCount   int    `json:"live_usage_count"`
	ObservedAt       string `json:"observed_at"`
	ToolLimitation   string `json:"tool_limitation"`
}

type figmaSourceAPIGeneratedAnnotationLiveScan struct {
	ObservedAt                   string                                             `json:"observed_at"`
	Tool                         string                                             `json:"tool"`
	PageCounts                   []figmaSourceAPIGeneratedAnnotationLivePageCounter `json:"page_counts"`
	TotalRiidoAnnotations        int                                                `json:"total_riido_annotations"`
	TotalAPIGeneratedAnnotations int                                                `json:"total_api_generated_annotations"`
}

type figmaSourceAPIGeneratedAnnotationLivePageCounter struct {
	PageID               string `json:"page_id"`
	PageName             string `json:"page_name"`
	RiidoAnnotationCount int    `json:"riido_annotation_count"`
	APIGeneratedCount    int    `json:"api_generated_count"`
	MissingOperationKind int    `json:"missing_operation_kind"`
	MissingBackground    int    `json:"missing_background"`
}
