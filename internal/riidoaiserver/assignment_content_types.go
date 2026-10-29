package riidoaiserver

import (
	"context"
	"encoding/json"
)

// AssignmentContentPart is an internal snapshot part, not a new public audit enum.
type AssignmentContentPart string

const (
	AssignmentContentMetadata    AssignmentContentPart = "metadata"
	AssignmentContentEditor      AssignmentContentPart = "editor"
	AssignmentContentDescription AssignmentContentPart = "description"
	AssignmentContentAlternate   AssignmentContentPart = "alternate"
	AssignmentContentComments    AssignmentContentPart = "comments"
	AssignmentContentBoard       AssignmentContentPart = "board"
)

// AssignmentContentManifest grants access only to pages actually captured.
// Incomplete parts may have no cursor; no cursor promises uncaptured content.
type AssignmentContentManifest struct {
	ResourceKind AssignmentContentKind `json:"resource_kind"`
	SourceID     string                `json:"source_id"`
	Part         AssignmentContentPart `json:"part"`
	Version      string                `json:"version"`
	Status       string                `json:"status"`
	Complete     bool                  `json:"complete"`
	Cursors      []string              `json:"cursors"`
	NextAction   string                `json:"next_action,omitempty"`
}

type AssignmentContentPage struct {
	ResourceKind AssignmentContentKind     `json:"resource_kind"`
	Cursor       string                    `json:"cursor"`
	SourceID     string                    `json:"source_id"`
	Part         AssignmentContentPart     `json:"part"`
	Version      string                    `json:"version"`
	Offset       int                       `json:"offset"`
	Status       string                    `json:"status"`
	Complete     bool                      `json:"complete"`
	Data         json.RawMessage           `json:"data"`
	Sources      []AssignmentContextSource `json:"sources"`
}

// AssignmentContentStore publishes one manifest and all payloads atomically.
// Existing legacy snapshots need not implement this additive capability.
type AssignmentContentStore interface {
	SaveAssignmentContentBundle(context.Context, AssignmentContextSnapshot, []AssignmentContentPage) error
	LoadAssignmentContentPage(context.Context, string, int) (AssignmentContentPage, bool, error)
}
