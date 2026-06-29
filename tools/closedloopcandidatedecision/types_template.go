package main

type decisionTemplate struct {
	SubjectKind  string `json:"subject_kind"`
	Disposition  string `json:"disposition"`
	Priority     string `json:"priority"`
	Owner        string `json:"owner"`
	NextLoop     string `json:"next_loop"`
	NextArtifact string `json:"next_artifact"`
	ReviewBy     string `json:"review_by,omitempty"`
	Reason       string `json:"reason"`
}
