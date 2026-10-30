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
	// OwnerKind/OwnerID name the work, meeting or Agent list a link, attachment,
	// comment or Agent profile belongs to. Omitted when the record is its own group.
	OwnerKind string `json:"owner_kind,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	// Listed is false for title/summary-only prompt evidence the read UI omits.
	Listed bool `json:"listed"`
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
	// SourceTotal counts the viewer-visible groups (owner, else the record
	// itself) that hold at least one listed record, across all pages.
	SourceTotal int `json:"source_total"`
}
