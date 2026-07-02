package main

import "encoding/json"

type report struct {
	SchemaVersion string                `json:"schema_version"`
	CapturedAt    string                `json:"captured_at"`
	Redacted      bool                  `json:"redacted"`
	Source        sourceSummary         `json:"source"`
	Endpoints     []endpointObservation `json:"endpoints"`
	V3            threadSummary         `json:"v3_threads"`
	V2            threadSummary         `json:"v2_threads"`
	Subscription  subscriptionSummary   `json:"thread_stream_subscription"`
	SSEEvents     []sseEventSummary     `json:"sse_events"`
	Decision      decisionSummary       `json:"decision"`
}

type sourceSummary struct {
	BaseURL        string `json:"base_url"`
	WorkspaceID    string `json:"workspace_id"`
	TaskID         string `json:"task_id"`
	ConversationID string `json:"conversation_id,omitempty"`
	TokenEnv       string `json:"token_env"`
}

type endpointObservation struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

type threadCollection struct {
	Threads      []threadRecord  `json:"threads"`
	ActiveStream json.RawMessage `json:"active_stream,omitempty"`
}

type threadRecord struct {
	ThreadID        string          `json:"thread_id"`
	ConversationID  string          `json:"conversation_id"`
	AssignmentID    string          `json:"assignment_id"`
	RunID           string          `json:"run_id"`
	WorkStatus      string          `json:"work_status"`
	AssignmentState string          `json:"assignment_state"`
	CommentKind     string          `json:"comment_kind"`
	ActiveStream    json.RawMessage `json:"active_stream,omitempty"`
	Messages        []messageRecord `json:"messages,omitempty"`
	Lines           []lineRecord    `json:"lines,omitempty"`
}
