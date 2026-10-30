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
	snapshot, err := s.commentMentionPrecheckSnapshot(ctx, current)
	if err != nil {
		return nil, err
	}
	call, principal := snapshot.call, snapshot.principal
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
	request := AssignRequest{AgentID: target, WorkspaceID: current.WorkspaceID, RuntimeProvider: binding.RuntimeProvider, ModelID: selected.ModelID}
	composed, err := composeAssignRequestWithTaskContextResult(current.TaskID, current.TaskID, request, snapshot.evidence.Context)
	if err != nil {
		return nil, err
	}
	request = composed.Request
	evidence := snapshot.evidence
	evidence.AgentID = target
	publication := AssignAIAgentTaskRequest{}
	// Use the same freshly verified task and root membership as execution.
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
