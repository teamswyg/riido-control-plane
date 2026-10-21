package main

const pendingReviewRef = defaultManifest + "#review_evidence/"

type pendingReview struct {
	SchemaVersion string             `json:"schema_version"`
	ReviewedOn    string             `json:"reviewed_on"`
	ReviewBy      string             `json:"review_by"`
	IntervalDays  int                `json:"review_interval_days"`
	Sources       []reviewSource     `json:"source_reviews"`
	Items         []reviewedDecision `json:"items"`
}

type reviewSource struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type reviewedDecision struct {
	decisionRecord
	EvidenceRefs []string `json:"evidence_refs"`
	Observation  string   `json:"observation"`
	RemainingGap string   `json:"remaining_gap"`
	NextStep     string   `json:"next_step"`
}
