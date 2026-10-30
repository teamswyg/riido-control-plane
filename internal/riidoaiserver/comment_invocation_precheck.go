package riidoaiserver

import (
	"context"
	"errors"
	"slices"
)

// The directory and terminal dispatch share these limits. A directory read is
// advisory and never reserves a slot; dispatch rechecks the whole fanout.
func commentMentionLimitReasons(call *CommentInvocation, target string, count int) []string {
	if call == nil {
		return []string{"comment_context_missing"}
	}
	var reasons []string
	if call.Depth >= 3 {
		reasons = append(reasons, "depth_limit")
	}
	if count > call.Remaining {
		reasons = append(reasons, "budget_exhausted")
	}
	if target != "" && slices.Contains(call.Ancestors, target) {
		reasons = append(reasons, "cycle")
	}
	return reasons
}

// precheckCommentMention runs for the authenticated assignment being read by
// the CLI. Reuse the same conversation/native-session preparation as dispatch,
// without publishing an assignment or consuming its remaining call budget.
func (s Server) precheckCommentMention(ctx context.Context, current Assignment, target string) ([]string, error) {
	principal, err := assignmentRequesterPrincipal(current, false)
	if err != nil {
		return nil, err
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, current.TaskID)
	if err != nil {
		return nil, err
	}
	var call *CommentInvocation
	for _, thread := range threads.Threads {
		if thread.AssignmentID != current.ID {
			continue
		}
		if thread.TaskID != current.TaskID || thread.AgentID != current.AgentID {
			return nil, ErrAuthorizationForbidden
		}
		call = thread.CommentInvocation
		if call != nil && (call.PrincipalID != current.CreatedBy || call.WorkspaceID != current.WorkspaceID || thread.ConversationID != "comment:"+call.RootCommentID) {
			return nil, ErrAuthorizationForbidden
		}
		break
	}
	if reasons := commentMentionLimitReasons(call, target, 1); len(reasons) > 0 {
		return reasons, nil
	}
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, current.TaskID)
	if err != nil {
		return nil, err
	}
	selected, err := selectAssignableAgent(agents.Agents, target)
	if err != nil {
		return []string{"target_unavailable"}, nil
	}
	binding, _, err := s.resolveAgentRuntimeFact(target)
	if err != nil {
		return []string{AgentDirectoryReasonRuntimeOffline}, nil
	}
	request := AssignRequest{AgentID: target, WorkspaceID: current.WorkspaceID, RuntimeProvider: binding.RuntimeProvider, ModelID: selected.ModelID, Worktree: current.Worktree}
	evidence := verifiedCommentDelivery{TaskID: current.TaskID, AgentID: target, RootCommentID: call.RootCommentID}
	publication := AssignAIAgentTaskRequest{}
	// Starting the current comment invocation already adopted legacy rows from
	// verified root membership, so this read uses that stored conversation ID.
	err = s.prepareCommentConversation(ctx, principal, evidence, &publication, &request, selected.OwnerPrincipalID)
	switch {
	case errors.Is(err, ErrAIAgentTaskThreadConflict), errors.Is(err, errNativeSessionUnavailable):
		return []string{"session_conflict"}, nil
	case errors.Is(err, ErrAuthorizationForbidden), errors.Is(err, ErrAIAgentNotFound):
		return []string{"target_unavailable"}, nil
	case err != nil:
		return nil, err
	default:
		return nil, nil
	}
}
