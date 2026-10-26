package riidoaiserver

import "context"

type preparedThreadMessage struct {
	Request            AssignRequest
	TaskContext        *AIAgentTaskContext
	StartsNewExecution bool
	Replay             *AIAgentTaskActionResponse
}

func (s Server) prepareThreadMessage(ctx context.Context, principal AuthorizationResult, bearer, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (preparedThreadMessage, error) {
	selected, history, err := s.readTaskThreadMessageContext(ctx, principal, taskID, threadID, req)
	if err != nil {
		return preparedThreadMessage{}, err
	}
	// Authorization and thread identity are checked before replay, but mutable task
	// documents are deliberately not read again after an execution was committed.
	replay, found, err := s.findTaskRequestReceipt(ctx, principal, taskID, AssignAIAgentTaskRequest{AgentID: selected.AgentID, RequestID: threadMessageRequestID(threadID, req.RequestID)})
	if err != nil {
		return preparedThreadMessage{}, err
	}
	if found {
		return preparedThreadMessage{Replay: &replay}, nil
	}
	request, taskContext, err := s.assignRequestFromAIAgentTaskThreadMessage(ctx, principal, bearer, taskID, threadID, req, selected, history)
	return preparedThreadMessage{Request: request, TaskContext: taskContext, StartsNewExecution: taskThreadMessageStartsNewExecution(AIAgentTaskThreadRecord{WorkStatus: selected.WorkStatus, AssignmentState: selected.AssignmentState})}, err
}
