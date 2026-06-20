package riidoaiserver

import "time"

type assignmentEventInput struct {
	AgentID      string
	TaskID       string
	AssignmentID string
	State        AssignmentState
	Type         string
	Message      string
	Metadata     map[string]string
	At           time.Time
}

func (input assignmentEventInput) IsProgressLog() bool {
	return input.Type == EventRiidoLog && input.Message != ""
}
