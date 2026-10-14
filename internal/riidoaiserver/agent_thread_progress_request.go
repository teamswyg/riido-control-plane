package riidoaiserver

import (
	"errors"
	"strings"
)

func normalizeAgentThreadProgressRequest(req *AgentThreadProgressBatchRequest) error {
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	req.RunID = strings.TrimSpace(req.RunID)
	if req.AssignmentID == "" {
		return errors.New("assignment_id is required")
	}
	if req.TaskID == "" {
		return errors.New("task_id is required")
	}
	if req.RunID == "" {
		req.RunID = "run-" + req.AssignmentID
	}
	req.Lines = normalizeProgressLines(req.Lines)
	if len(req.Lines) == 0 {
		return errors.New("lines are required")
	}
	return nil
}
