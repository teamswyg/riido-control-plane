package main

type evidence struct {
	SchemaVersion      string         `json:"schema_version"`
	Status             string         `json:"status"`
	GeneratedAt        string         `json:"generated_at"`
	ExpiresAt          string         `json:"expires_at"`
	EvidenceTTLHours   int            `json:"evidence_ttl_hours"`
	LoopRegistry       string         `json:"loop_registry_manifest"`
	CheckCount         int            `json:"check_count"`
	MeasurementCount   int            `json:"measurement_count"`
	CoveredCount       int            `json:"covered_count"`
	PartialCount       int            `json:"partial_count"`
	StalePartialCount  int            `json:"stale_partial_count"`
	StaleAfterDays     int            `json:"stale_after_days"`
	RequiredCategories []string       `json:"required_categories"`
	MissingCategories  []string       `json:"missing_categories"`
	CategoryCounts     map[string]int `json:"category_counts"`
	MeasurementKinds   map[string]int `json:"measurement_kinds"`
	StatusCounts       map[string]int `json:"status_counts"`
	PartialChecks      []partialCheck `json:"partial_checks"`
	Loop               loopSpec       `json:"loop"`
}

type partialCheck struct {
	ID           string `json:"id"`
	Date         string `json:"date"`
	Category     string `json:"category"`
	AgeDays      int    `json:"age_days"`
	Stale        bool   `json:"stale"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command"`
}
