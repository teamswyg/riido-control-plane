package main

type subscriptionPayload struct {
	Stream struct {
		Href      string `json:"href"`
		EventType string `json:"event_type"`
	} `json:"stream"`
	ActiveThreadFilters []threadFilter `json:"active_thread_filters"`
}

type threadFilter struct {
	ThreadID string `json:"thread_id"`
	RunID    string `json:"run_id"`
	AgentID  string `json:"agent_id"`
}

type subscriptionSummary struct {
	StreamHref               string `json:"stream_href"`
	EventType                string `json:"event_type"`
	ActiveThreadFilterCount  int    `json:"active_thread_filter_count"`
	HighlightedFilterMatched bool   `json:"highlighted_filter_matched"`
}

type sseEventSummary struct {
	Event           string `json:"event"`
	TaskID          string `json:"task_id,omitempty"`
	ThreadID        string `json:"thread_id,omitempty"`
	ConversationID  string `json:"conversation_id,omitempty"`
	AssignmentID    string `json:"assignment_id,omitempty"`
	RunID           string `json:"run_id,omitempty"`
	WorkStatus      string `json:"work_status,omitempty"`
	AssignmentState string `json:"assignment_state,omitempty"`
	LineCount       int    `json:"line_count"`
}

type sseProgressPayload struct {
	TaskID          string       `json:"task_id"`
	ThreadID        string       `json:"thread_id"`
	ConversationID  string       `json:"conversation_id"`
	AssignmentID    string       `json:"assignment_id"`
	RunID           string       `json:"run_id"`
	WorkStatus      string       `json:"work_status"`
	AssignmentState string       `json:"assignment_state"`
	Lines           []lineRecord `json:"lines"`
}
