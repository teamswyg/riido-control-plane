package main

type loopSurface struct {
	ID                string           `json:"id"`
	Kind              string           `json:"kind"`
	Observes          []string         `json:"observes"`
	Verifies          []string         `json:"verifies"`
	Evidence          []evidenceSource `json:"evidence"`
	RefreshWorkflow   string           `json:"refresh_workflow"`
	ExpiresAfterHours int              `json:"expires_after_hours"`
	FailsWhen         []string         `json:"fails_when"`
	PromotesTo        []string         `json:"promotes_to,omitempty"`
	Providers         []string         `json:"providers,omitempty"`
}
