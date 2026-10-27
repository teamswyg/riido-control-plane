package riidoaiserver

import (
	"context"
	"time"
)

const aiAgentRunReadsPageSize = 100

type aiAgentRunReader interface {
	FindAIAgentTaskThreadByRunID(context.Context, string, string) (AIAgentTaskThreadRecord, error)
}

type AIAgentRunRead struct {
	AgentContextReadSource
	Via               string    `json:"via"`
	ServedAt          time.Time `json:"served_at"`
	Title             string    `json:"title,omitempty"`
	ModifiedSinceRead *bool     `json:"modified_since_read,omitempty"`
}

type AIAgentRunReadsResponse struct {
	SchemaVersion string           `json:"schema_version"`
	RunID         string           `json:"run_id"`
	TaskID        string           `json:"task_id"`
	ThreadID      string           `json:"thread_id"`
	AssignmentID  string           `json:"assignment_id"`
	Reads         []AIAgentRunRead `json:"reads"`
	Page          int              `json:"page"`
	PageSize      int              `json:"page_size"`
	HasMore       bool             `json:"has_more"`
}
