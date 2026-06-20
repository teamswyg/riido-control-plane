package riidoaiserver

import (
	"errors"
	"strings"
)

func newThreadProgressInput(agentID string, req AgentThreadProgressBatchRequest) (threadProgressInput, error) {
	input := threadProgressInput{AgentID: strings.TrimSpace(agentID), Request: req}
	input.Request.TaskID = strings.TrimSpace(input.Request.TaskID)
	input.Request.ThreadID = strings.TrimSpace(input.Request.ThreadID)
	input.Request.AssignmentID = strings.TrimSpace(input.Request.AssignmentID)
	input.Request.RunID = strings.TrimSpace(input.Request.RunID)
	if input.AgentID == "" {
		return threadProgressInput{}, errors.New("agent_id is required")
	}
	if input.Request.TaskID == "" {
		return threadProgressInput{}, errors.New("task_id is required")
	}
	if input.Request.AssignmentID == "" {
		return threadProgressInput{}, errors.New("assignment_id is required")
	}
	input.Lines = normalizeProgressLines(input.Request.Lines)
	if len(input.Lines) == 0 {
		return threadProgressInput{}, errors.New("lines are required")
	}
	if input.Request.RunID == "" {
		input.Request.RunID = "run-" + input.Request.AssignmentID
	}
	if input.Request.ThreadID == "" {
		input.Request.ThreadID = threadIDForRun(input.Request.TaskID, input.AgentID, input.Request.RunID)
		input.GeneratedThreadID = true
	}
	return input, nil
}
