package riidoaiserver

import "context"

func (s Server) assignThreadMessage(ctx context.Context, taskID string, req AssignRequest, startsNewExecution bool) (Assignment, error) {
	if startsNewExecution && req.RequestID != "" {
		return s.assignment.AssignTaskAdditive(ctx, taskID, req)
	}
	return s.assignment.AssignTask(ctx, taskID, req)
}

func taskThreadMessageReceiptKey(principal AuthorizationResult, taskID, threadID, agentID string, req CreateAIAgentTaskThreadMessageRequest) string {
	if req.toolApproval || req.toolApprovalWithoutPending {
		return ""
	}
	return taskRequestReceiptKey(principal, taskID, AssignAIAgentTaskRequest{AgentID: agentID, RequestID: threadMessageRequestID(threadID, req.RequestID)})
}
