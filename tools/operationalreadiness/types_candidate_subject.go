package main

type candidateSubject struct {
	Kind           string `json:"kind"`
	CheckID        string `json:"check_id"`
	Category       string `json:"category"`
	AgeDays        int    `json:"age_days"`
	Stale          bool   `json:"stale"`
	NextArtifact   string `json:"next_artifact"`
	NextCommand    string `json:"next_command"`
	StaleAfterDays int    `json:"stale_after_days"`
}
