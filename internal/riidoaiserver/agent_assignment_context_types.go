package riidoaiserver

import (
	"strings"
	"time"
)

type AgentAssignmentContextRequest struct {
	AssignmentID string `json:"assignment_id"`
	LeaseToken   string `json:"lease_token"`
	View         string `json:"view"`
	Target       string `json:"target,omitempty"`
	Cursor       string `json:"cursor,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	// MaxBytes asks a content read for consecutive pages of the same part up to
	// this encoded size. Zero keeps the single-page response.
	MaxBytes int `json:"max_bytes,omitempty"`
}

type AgentAssignmentContextResponse struct {
	SchemaVersion string                    `json:"schema_version"`
	AssignmentID  string                    `json:"assignment_id"`
	TaskID        string                    `json:"task_id"`
	View          string                    `json:"view"`
	CapturedAt    time.Time                 `json:"captured_at"`
	Truncated     bool                      `json:"truncated,omitempty"`
	Live          bool                      `json:"live,omitempty"`
	Data          any                       `json:"data"`
	Sources       []AssignmentContextSource `json:"sources"`
}

// normalizeAssignmentContextRequest defaults the view to base and reports
// whether the request names an assignment, a lease token and a known view.
func normalizeAssignmentContextRequest(req *AgentAssignmentContextRequest) bool {
	req.View = strings.TrimSpace(req.View)
	if req.View == "" {
		req.View = AssignmentContextViewBase
	}
	return strings.TrimSpace(req.AssignmentID) != "" && strings.TrimSpace(req.LeaseToken) != "" && validAssignmentContextView(req.View) && validAssignmentContentBatchBytes(req.View, req.MaxBytes)
}
