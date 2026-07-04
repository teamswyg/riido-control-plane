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

type architectureFileEvidence struct {
	Path                   string   `json:"path"`
	ComponentIDs           []string `json:"component_ids"`
	HotPathIDs             []string `json:"hot_path_ids"`
	HotPathCategories      []string `json:"hot_path_categories"`
	PressureDimensions     []string `json:"pressure_dimensions"`
	ObservabilitySignals   []string `json:"observability_signals"`
	EvidenceRefs           []string `json:"evidence_refs"`
	Benchmarks             []string `json:"benchmarks,omitempty"`
	Tests                  []string `json:"tests,omitempty"`
	OptimizationCandidates []string `json:"optimization_candidates,omitempty"`
	TargetVerifierCommands []string `json:"target_verifier_commands,omitempty"`
}

type pressureDimensionEvidence struct {
	Dimension              string   `json:"dimension"`
	ComponentIDs           []string `json:"component_ids"`
	Files                  []string `json:"files"`
	HotPathIDs             []string `json:"hot_path_ids"`
	ObservabilitySignals   []string `json:"observability_signals"`
	EvidenceRefs           []string `json:"evidence_refs"`
	TargetVerifierCommands []string `json:"target_verifier_commands"`
}
