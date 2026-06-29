package main

type evidence struct {
	SchemaVersion      string         `json:"schema_version"`
	Status             string         `json:"status"`
	CheckCount         int            `json:"check_count"`
	CoveredCount       int            `json:"covered_count"`
	PartialCount       int            `json:"partial_count"`
	RequiredCategories []string       `json:"required_categories"`
	MissingCategories  []string       `json:"missing_categories"`
	CategoryCounts     map[string]int `json:"category_counts"`
	StatusCounts       map[string]int `json:"status_counts"`
	PartialChecks      []partialCheck `json:"partial_checks"`
	Loop               loopSpec       `json:"loop"`
}

type partialCheck struct {
	ID           string `json:"id"`
	Category     string `json:"category"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command"`
}
