package main

type notionOpenLoop struct {
	PageID          string        `json:"page_id"`
	PageTitle       string        `json:"page_title"`
	PageURL         string        `json:"page_url"`
	CapturedAt      string        `json:"captured_at"`
	RefreshWorkflow string        `json:"refresh_workflow"`
	CadenceHours    int           `json:"cadence_hours"`
	StatusTags      []string      `json:"status_tags"`
	Cycles          []notionCycle `json:"cycles"`
}

type notionCycle struct {
	ID                   string        `json:"id"`
	Priority             string        `json:"priority"`
	Status               string        `json:"status"`
	CodexStatus          string        `json:"codex_status"`
	Source               string        `json:"source"`
	Summary              string        `json:"summary"`
	BackfilledCheck      string        `json:"backfilled_check"`
	RequiredNextArtifact string        `json:"required_next_artifact"`
	RequiredNextCommand  string        `json:"required_next_command"`
	EvidenceRefs         []evidenceRef `json:"evidence_refs"`
}

type notionEvidence struct {
	PageTitle     string                `json:"page_title"`
	PageURL       string                `json:"page_url"`
	CapturedAt    string                `json:"captured_at"`
	CadenceHours  int                   `json:"cadence_hours"`
	CycleCount    int                   `json:"cycle_count"`
	P0Count       int                   `json:"p0_count"`
	PartialCount  int                   `json:"partial_count"`
	CoveredCount  int                   `json:"covered_count"`
	StatusCounts  map[string]int        `json:"status_counts"`
	CodexStatuses map[string]int        `json:"codex_statuses"`
	Cycles        []notionCycleEvidence `json:"cycles"`
}

type notionCycleEvidence struct {
	ID                   string `json:"id"`
	Priority             string `json:"priority"`
	Status               string `json:"status"`
	CodexStatus          string `json:"codex_status"`
	Source               string `json:"source"`
	BackfilledCheck      string `json:"backfilled_check"`
	RequiredNextArtifact string `json:"required_next_artifact"`
}
