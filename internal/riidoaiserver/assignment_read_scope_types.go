package riidoaiserver

// AssignmentReadScope stores authorized selection, not a live-query capability.
// The complete captured values are addressed by the separately stored parts.
type AssignmentReadScope struct {
	SchemaVersion string                     `json:"schema_version"`
	RootID        string                     `json:"root_id"`
	TeamID        string                     `json:"team_id"`
	Status        string                     `json:"status"`
	Resources     []AssignmentScopedResource `json:"resources"`
}

type AssignmentScopedResource struct {
	Kind      AssignmentContentKind       `json:"kind"`
	ID        string                      `json:"id"`
	TeamID    string                      `json:"team_id"`
	Title     string                      `json:"title"`
	KeyNumber string                      `json:"key_number,omitempty"`
	UpdatedAt string                      `json:"updated_at,omitempty"`
	Reasons   []taskReadScopeReason       `json:"reasons"`
	Parts     []AssignmentContentManifest `json:"parts"`
}

type taskReadScopeReason struct {
	Source   *string `json:"source,omitempty"`
	Kind     string  `json:"kind"`
	Relation *string `json:"relation"`
}

// Ephemeral capture state is carried from the authenticated read to the
// atomic assignment snapshot write. It never contains credentials.
type assignmentReadScopeCapture struct {
	Manifest []AssignmentContentManifest
	Pages    []AssignmentContentPage
}
