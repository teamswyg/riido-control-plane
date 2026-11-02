package riidoaiserver

type AssignmentContextChanges struct {
	SnapshotID          string                    `json:"snapshot_id"`
	Since               string                    `json:"since"`
	Changes             []AssignmentContextChange `json:"changes"`
	Complete            bool                      `json:"complete"`
	FullRefreshRequired bool                      `json:"full_refresh_required"`
	NextAction          string                    `json:"next_action,omitempty"`
}

type AssignmentContextChange struct {
	ResourceKind AssignmentContentKind `json:"resource_kind"`
	SourceID     string                `json:"source_id"`
	Part         AssignmentContentPart `json:"part"`
	Status       string                `json:"status"`
	Reason       string                `json:"reason,omitempty"`
	// Metadata is present only for a resource in the current authorized capture.
	Title     string                     `json:"title,omitempty"`
	TeamID    string                     `json:"team_id,omitempty"`
	KeyNumber string                     `json:"key_number,omitempty"`
	Reasons   []taskReadScopeReason      `json:"reasons,omitempty"`
	Current   *AssignmentContentManifest `json:"current,omitempty"`
}
