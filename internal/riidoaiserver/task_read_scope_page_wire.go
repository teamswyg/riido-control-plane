package riidoaiserver

import "encoding/json"

type taskReadBodyPage struct {
	Selector   *taskReadPartSelector `json:"selector,omitempty"`
	ResourceID string                `json:"resourceId"`
	Part       string                `json:"part"`
	Status     string                `json:"status"`
	Source     string                `json:"source"`
	Format     string                `json:"format"`
	Version    string                `json:"version"`
	Offset     *int                  `json:"offset"`
	Content    string                `json:"content"`
	NextOffset *int                  `json:"nextOffset"`
	Complete   *bool                 `json:"complete"`
}

type taskReadCommentPage struct {
	ContentPage     *taskReadBodyPage     `json:"contentPage,omitempty"`
	ResourceID      string                `json:"resourceId"`
	ResourceKind    AssignmentContentKind `json:"resourceKind"`
	ParentCommentID *string               `json:"parentCommentId"`
	Version         string                `json:"version"`
	Status          string                `json:"status"`
	Page            int                   `json:"page"`
	Limit           int                   `json:"limit"`
	NextPage        *int                  `json:"nextPage"`
	Complete        *bool                 `json:"complete"`
	Items           []json.RawMessage     `json:"items"`
}

type taskReadMeeting struct {
	Summary         *taskReadResourceSummary `json:"summary,omitempty"`
	ContentPage     *taskReadBodyPage        `json:"contentPage,omitempty"`
	ID              string                   `json:"id"`
	TeamID          string                   `json:"teamId"`
	Status          string                   `json:"status"`
	Details         json.RawMessage          `json:"details"`
	Labels          []json.RawMessage        `json:"labels"`
	Board           *taskReadBoardPage       `json:"board"`
	Body            *taskReadBodyPage        `json:"body"`
	DescriptionBody *taskReadBodyPage        `json:"descriptionBody"`
	AlternateBody   *taskReadBodyPage        `json:"alternateBody"`
	PartsStatus     []*taskReadPartStatus    `json:"partsStatus"`
}

type taskReadBoardPage struct {
	ContentPage *taskReadBodyPage `json:"contentPage,omitempty"`
	Version     string            `json:"version"`
	Offset      *int              `json:"offset"`
	NextCursor  *string           `json:"nextCursor"`
	Complete    *bool             `json:"complete"`
	Groups      []json.RawMessage `json:"groups"`
	Cards       []json.RawMessage `json:"cards"`
}
