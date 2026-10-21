package riidoaiserver

import (
	"context"
	"errors"
)

var errThreadMessageExecutionCancelled = errors.New("thread message execution was cancelled before publication")

func (s Server) assignThreadMessage(ctx context.Context, taskID string, req AssignRequest, startsNewExecution bool) (Assignment, error) {
	if startsNewExecution && req.RequestID != "" {
		assignment, err := s.assignment.AssignTaskAdditive(ctx, taskID, req)
		if err == nil && (assignment.State == AssignmentCancelled || assignment.State == AssignmentCancelling) {
			return Assignment{}, errThreadMessageExecutionCancelled
		}
		return assignment, err
	}
	return s.assignment.AssignTask(ctx, taskID, req)
}

func taskThreadMessageReceiptKey(principal AuthorizationResult, taskID, threadID, agentID string, req CreateAIAgentTaskThreadMessageRequest) string {
	if req.toolApproval || req.toolApprovalWithoutPending {
		return ""
	}
	return taskRequestReceiptKey(principal, taskID, AssignAIAgentTaskRequest{AgentID: agentID, RequestID: threadMessageRequestID(threadID, req.RequestID)})
}
