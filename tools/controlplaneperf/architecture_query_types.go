package main

type architectureQueryEvidence struct {
	SchemaVersion string                 `json:"schema_version"`
	Status        string                 `json:"status"`
	QueryCount    int                    `json:"query_count"`
	HitCount      int                    `json:"hit_count"`
	MissCount     int                    `json:"miss_count"`
	Queries       []architectureQueryRow `json:"queries"`
}

type architectureQueryRow struct {
	Path                   string   `json:"path"`
	Matched                bool     `json:"matched"`
	ComponentIDs           []string `json:"component_ids,omitempty"`
	HotPathIDs             []string `json:"hot_path_ids,omitempty"`
	PressureDimensions     []string `json:"pressure_dimensions,omitempty"`
	ObservabilitySignals   []string `json:"observability_signals,omitempty"`
	TargetVerifierCommands []string `json:"target_verifier_commands,omitempty"`
	OptimizationCandidates []string `json:"optimization_candidates,omitempty"`
}
