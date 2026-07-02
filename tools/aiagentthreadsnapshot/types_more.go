package main

type messageRecord struct {
	Role          string `json:"role"`
	CommentKind   string `json:"comment_kind"`
	AssignmentID  string `json:"assignment_id"`
	RunID         string `json:"run_id"`
	Seq           int    `json:"seq"`
	Body          string `json:"body"`
	ResultMessage string `json:"result_message"`
}

type lineRecord struct {
	Seq     int    `json:"seq"`
	Message string `json:"message"`
}

type threadSummary struct {
	ThreadCount               int             `json:"thread_count"`
	RunningCount              int             `json:"running_count"`
	QueuedCount               int             `json:"queued_count"`
	TerminalCount             int             `json:"terminal_count"`
	TerminalActiveStreamCount int             `json:"terminal_active_stream_count"`
	ActiveStream              bool            `json:"active_stream"`
	HighlightedThreads        []threadSurface `json:"highlighted_threads"`
}

type threadSurface struct {
	ThreadID        string `json:"thread_id"`
	ConversationID  string `json:"conversation_id,omitempty"`
	AssignmentID    string `json:"assignment_id,omitempty"`
	RunID           string `json:"run_id,omitempty"`
	WorkStatus      string `json:"work_status"`
	AssignmentState string `json:"assignment_state"`
	CommentKind     string `json:"comment_kind,omitempty"`
	MessageCount    int    `json:"message_count,omitempty"`
	LineCount       int    `json:"line_count,omitempty"`
	ActiveStream    bool   `json:"active_stream"`
}

type decisionSummary struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type conversationCandidate struct {
	ConversationID string `json:"conversation_id"`
	ThreadCount    int    `json:"thread_count"`
	RunningCount   int    `json:"running_count"`
	QueuedCount    int    `json:"queued_count"`
	TerminalCount  int    `json:"terminal_count"`
	ActiveStreams  int    `json:"active_stream_count"`
	ThreadID       string `json:"sample_thread_id,omitempty"`
	AssignmentID   string `json:"sample_assignment_id,omitempty"`
	RunID          string `json:"sample_run_id,omitempty"`
}
