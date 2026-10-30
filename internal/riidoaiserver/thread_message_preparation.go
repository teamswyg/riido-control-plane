package riidoaiserver

import "context"

type preparedThreadMessage struct {
	CommentInvocation  *CommentInvocation
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
	target := threadMessageExecutionThread(selected, history, req.targetAgentID)
	// Authorization and thread identity are checked before replay, but mutable task
	// documents are deliberately not read again after an execution was committed.
	replay, found, err := s.findTaskRequestReceipt(ctx, principal, taskID, AssignAIAgentTaskRequest{AgentID: target.AgentID, RequestID: threadMessageRequestID(threadID, req.RequestID)})
	if err != nil {
		return preparedThreadMessage{}, err
	}
	if found {
		return preparedThreadMessage{Replay: &replay}, nil
	}
	if target.AgentID != selected.AgentID && target.ThreadID != "" && !taskThreadMessageStartsNewExecution(target) {
		return preparedThreadMessage{}, ErrAIAgentTaskThreadConflict
	}
	request, taskContext, err := s.assignRequestFromAIAgentTaskThreadMessage(ctx, principal, bearer, taskID, threadID, req, selected, history)
	prepared := preparedThreadMessage{Request: request, TaskContext: taskContext, StartsNewExecution: target.AgentID != selected.AgentID || taskThreadMessageStartsNewExecution(selected)}
	if err == nil && prepared.StartsNewExecution {
		prepared.CommentInvocation, err = s.prepareCommentFollowup(ctx, principal, selected, target.AgentID)
		if err == nil && prepared.CommentInvocation != nil {
			err = s.appendCommentInvocationPrompt(ctx, principal, taskID, &prepared.Request)
		}
	}
	return prepared, err
}
