package riidoaiserver

import "time"

type AgentDashboardTask struct {
	CreatedBy          string     `json:"created_by,omitempty"`
	AgentID            string     `json:"agent_id"`
	TaskID             string     `json:"task_id"`
	AssignmentID       string     `json:"assignment_id"`
	RunID              string     `json:"run_id"`
	Title              string     `json:"title"`
	KeyNumber          string     `json:"key_number"`
	State              string     `json:"state"`
	Message            string     `json:"message"`
	ConfirmationReason string     `json:"confirmation_reason,omitempty"`
	QueuePosition      int        `json:"queue_position"`
	QueuedAt           time.Time  `json:"queued_at"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	EndedAt            *time.Time `json:"ended_at,omitempty"`
	DurationSeconds    *int64     `json:"duration_seconds,omitempty"`
}

type AgentDashboardDay struct {
	Date      string `json:"date"`
	Completed int    `json:"completed"`
}

type AgentDashboardResponse struct {
	SchemaVersion     string               `json:"schema_version"`
	AgentID           string               `json:"agent_id"`
	WorkStatus        string               `json:"work_status"`
	Current           *AgentDashboardTask  `json:"current,omitempty"`
	Queue             []AgentDashboardTask `json:"queue"`
	WaitingCount      int                  `json:"waiting_count"`
	ConfirmationCount int                  `json:"confirmation_count"`
	CompletedRuns     int                  `json:"completed_runs"`
	History           []AgentDashboardTask `json:"history"`
	Page              int                  `json:"page"`
	PageSize          int                  `json:"page_size"`
	Total             int                  `json:"total"`
	HasNextPage       bool                 `json:"has_next_page"`
	Activity          []AgentDashboardDay  `json:"activity"`
	ActivityStart     string               `json:"activity_start"`
	ActivityEnd       string               `json:"activity_end"`
	ActivityCompleted int                  `json:"activity_completed"`
	ActiveDays        int                  `json:"active_days"`
	TimeZone          string               `json:"time_zone"`
	ObservedSince     *time.Time           `json:"observed_since,omitempty"`
	HistoryDurable    bool                 `json:"history_durable"`
	Scope             string               `json:"scope"`
}
