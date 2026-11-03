package riidoaiserver

import "context"

func (s Server) directAssignmentFollowup(ctx context.Context, p AuthorizationResult, selected AIAgentTaskThreadRecord, target string) (*CommentInvocation, error) {
	if s.config.CommentDelivery.Key == "" || selected.ThreadID == "" {
		return nil, nil
	}
	call := selected.CommentInvocation
	if call == nil {
		// Legacy conversations have no binding. Recover only from a verified original
		// direct assignment, never from a client-supplied thread ID or a past mention.
		rows, err := s.aiAgent.ListAIAgentTaskThreads(ctx, p, selected.TaskID)
		if err != nil {
			return nil, err
		}
		root, ok := verifiedDirectConversationRoot(selected, rows.Threads)
		if !ok {
			return nil, nil
		}
		call = root.CommentInvocation
		if call == nil {
			call, err = s.recoverDirectInvocation(ctx, p, root)
			if err != nil {
				return nil, err
			}
		}
	}
	if call == nil || call.DirectRoot == nil || call.PrincipalID != p.PrincipalID || call.WorkspaceID != p.WorkspaceID || !call.matchesConversation(selected) {
		return nil, nil
	}
	next := *call
	next.Ancestors = []string{target}
	next.Remaining = 8
	next.Depth = 0
	next.Outcomes = nil
	next.OriginAssignmentID = ""
	next.OriginMessageID = ""
	next.OriginThreadID = ""
	next.HumanFollowup = true
	return &next, nil
}
