package riidoaiserver

import (
	"errors"
	"strings"
)

func newAssignmentEventInput(agentID string, req AgentEventRequest, event TaskEvent) (assignmentEventInput, error) {
	input := assignmentEventInput{
		AgentID:      strings.TrimSpace(agentID),
		TaskID:       firstNonEmpty(event.TaskID, req.TaskID),
		AssignmentID: firstNonEmpty(event.AssignmentID, req.AssignmentID),
		State:        event.State,
		Type:         strings.TrimSpace(event.Type),
		Message:      firstNonEmpty(event.Message, req.Message),
		Metadata:     event.Metadata,
		At:           event.At,
	}
	if input.State == "" {
		input.State = req.State
	}
	if input.State == "" {
		input.State = AssignmentRunning
	}
	if input.AgentID == "" {
		return assignmentEventInput{}, errors.New("agent_id is required")
	}
	if input.TaskID == "" {
		return assignmentEventInput{}, errors.New("task_id is required")
	}
	if input.AssignmentID == "" {
		return assignmentEventInput{}, errors.New("assignment_id is required")
	}
	return input, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
