package main

type architectureQueryEvidence struct {
	SchemaVersion    string                 `json:"schema_version"`
	Status           string                 `json:"status"`
	QueryCount       int                    `json:"query_count"`
	HitCount         int                    `json:"hit_count"`
	DirectHitCount   int                    `json:"direct_hit_count"`
	FallbackHitCount int                    `json:"fallback_hit_count"`
	MissCount        int                    `json:"miss_count"`
	Queries          []architectureQueryRow `json:"queries"`
}

type architectureQueryRow struct {
	Path                   string                       `json:"path"`
	Matched                bool                         `json:"matched"`
	MatchKind              string                       `json:"match_kind"`
	ComponentIDs           []string                     `json:"component_ids,omitempty"`
	Components             []architectureQueryComponent `json:"components,omitempty"`
	HotPathIDs             []string                     `json:"hot_path_ids,omitempty"`
	PressureDimensions     []string                     `json:"pressure_dimensions,omitempty"`
	ObservabilitySignals   []string                     `json:"observability_signals,omitempty"`
	EvidenceRefs           []string                     `json:"evidence_refs,omitempty"`
	TargetVerifierCommands []string                     `json:"target_verifier_commands,omitempty"`
	OptimizationCandidates []string                     `json:"optimization_candidates,omitempty"`
}

type architectureQueryComponent struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}
