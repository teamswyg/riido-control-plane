package riidoaiserver

import "time"

const AgentClientEventTaskNotification = "agent_task_notification"

type AgentTaskNotificationEvent struct {
	TeamID               string               `json:"team_id,omitempty"`
	EventType            string               `json:"event_type"`
	SchemaVersion        string               `json:"schema_version"`
	EventID              string               `json:"event_id"`
	OccurredAt           time.Time            `json:"occurred_at"`
	WorkspaceID          string               `json:"workspace_id"`
	TaskID               string               `json:"task_id"`
	AgentID              string               `json:"agent_id"`
	AssignmentID         string               `json:"assignment_id"`
	RunID                string               `json:"run_id"`
	CreatedBy            string               `json:"created_by"`
	RecipientPrincipalID string               `json:"recipient_principal_id"`
	NotificationType     string               `json:"notification_type"`
	ConfirmationReason   string               `json:"confirmation_reason,omitempty"`
	WorkStatus           AgentWorkStatus      `json:"work_status"`
	AssignmentState      AgentAssignmentState `json:"assignment_state"`
}
