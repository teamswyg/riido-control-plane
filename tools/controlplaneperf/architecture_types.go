package main

type architectureComponent struct {
	ID                   string   `json:"id"`
	Role                 string   `json:"role"`
	HotPathCategories    []string `json:"hot_path_categories"`
	Files                []string `json:"files"`
	PressureDimensions   []string `json:"pressure_dimensions"`
	ObservabilitySignals []string `json:"observability_signals"`
	EvidenceRefs         []string `json:"evidence_refs"`
}

type architectureComponentEvidence struct {
	ID                   string   `json:"id"`
	HotPathCategories    []string `json:"hot_path_categories"`
	PressureDimensions   []string `json:"pressure_dimensions"`
	ObservabilitySignals []string `json:"observability_signals"`
	EvidenceRefs         []string `json:"evidence_refs"`
}
