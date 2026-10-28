package riidoaiserver

import "context"

func (s Server) commentInvocationSource(ctx context.Context, principal AuthorizationResult, event TaskEvent) (AIAgentTaskThreadRecord, error) {
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, event.TaskID)
	if err != nil {
		return AIAgentTaskThreadRecord{}, err
	}
	for _, thread := range threads.Threads {
		call := thread.CommentInvocation
		if thread.AssignmentID == event.AssignmentID && thread.AgentID == event.AgentID && call != nil && thread.AssignmentState == AgentAssignmentStateCompleted && thread.ConversationID == "comment:"+call.RootCommentID && call.PrincipalID == event.CreatedBy && call.WorkspaceID == event.WorkspaceID {
			return thread, nil
		}
	}
	return AIAgentTaskThreadRecord{}, nil
}

func commentInvocationEvidenceMatches(evidence verifiedCommentDelivery, taskID string, call *CommentInvocation) bool {
	root := evidence.RootCommentID
	if root == "" {
		root = evidence.CommentID
	}
	return evidence.TaskID == taskID && evidence.PrincipalID == call.PrincipalID && evidence.WorkspaceID == call.WorkspaceID && root == call.RootCommentID && len(call.Ancestors) > 0 && evidence.AgentID == call.Ancestors[0]
}
