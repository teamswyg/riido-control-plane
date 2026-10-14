package riidoaiserver

import (
	"errors"
	"log"
	"strings"
)

const (
	agentEventErrorCategoryAssignmentNotFound = "assignment_not_found"
	agentEventErrorCategoryAgentMismatch      = "assignment_agent_mismatch"
	agentEventErrorCategoryTaskMismatch       = "assignment_task_mismatch"
	agentEventErrorCategoryInvalidTransition  = "invalid_state_transition"
	agentEventErrorCategoryBindingValidation  = "binding_validation"
	agentEventErrorCategoryStoreFailure       = "store_failure"
	agentEventErrorCategoryBadRequest         = "bad_request"
)

func logAgentEventRejected(agentID string, req AgentEventRequest, err error) {
	log.Printf(
		"event=agent_event_rejected route=/v1/agents/{agent_id}/events status=400 agent_id=%q assignment_id=%q event_type=%q requested_state=%q event_error_category=%q",
		strings.TrimSpace(agentID),
		strings.TrimSpace(req.AssignmentID),
		strings.TrimSpace(req.EventType),
		strings.TrimSpace(string(req.State)),
		agentEventErrorCategory(err),
	)
}

func agentEventErrorCategory(err error) string {
	if err == nil {
		return agentEventErrorCategoryBadRequest
	}
	message := err.Error()
	switch {
	case strings.HasPrefix(message, "assignment ") && strings.HasSuffix(message, " not found"):
		return agentEventErrorCategoryAssignmentNotFound
	case strings.Contains(message, " belongs to agent "):
		return agentEventErrorCategoryAgentMismatch
	case strings.Contains(message, " belongs to task "):
		return agentEventErrorCategoryTaskMismatch
	case strings.Contains(message, "invalid assignment transition "):
		return agentEventErrorCategoryInvalidTransition
	case errors.Is(err, ErrAgentBindingValidation):
		return agentEventErrorCategoryBindingValidation
	case strings.Contains(message, "dynamodb ") ||
		strings.Contains(message, "save assignment") ||
		strings.Contains(message, "snapshot"):
		return agentEventErrorCategoryStoreFailure
	default:
		return agentEventErrorCategoryBadRequest
	}
}
