package riidoaiserver

import "context"

type commentMentionPrecheckCacheKey struct{}

type commentMentionPrecheckCache struct {
	loaded   bool
	snapshot commentMentionPrecheckSnapshot
	err      error
}

type commentMentionPrecheckSnapshot struct {
	principal AuthorizationResult
	call      *CommentInvocation
	evidence  verifiedCommentDelivery
}

// One directory request can check 50 candidates. Reuse its verified evidence,
// not a global cache: the next list/show and actual dispatch fetch it again.
func (s Server) commentMentionPrecheckSnapshot(ctx context.Context, current Assignment) (commentMentionPrecheckSnapshot, error) {
	cache, _ := ctx.Value(commentMentionPrecheckCacheKey{}).(*commentMentionPrecheckCache)
	if cache == nil {
		cache = &commentMentionPrecheckCache{}
	}
	if !cache.loaded {
		cache.loaded = true
		cache.snapshot, cache.err = s.readCommentMentionPrecheckSnapshot(ctx, current)
	}
	return cache.snapshot, cache.err
}

func (s Server) readCommentMentionPrecheckSnapshot(ctx context.Context, current Assignment) (commentMentionPrecheckSnapshot, error) {
	var out commentMentionPrecheckSnapshot
	principal, err := assignmentRequesterPrincipal(current, false)
	if err != nil {
		return out, err
	}
	out.principal = principal
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, current.TaskID)
	if err != nil {
		return out, err
	}
	for _, thread := range threads.Threads {
		if thread.AssignmentID != current.ID {
			continue
		}
		if thread.TaskID != current.TaskID || thread.AgentID != current.AgentID {
			return out, ErrAuthorizationForbidden
		}
		out.call = thread.CommentInvocation
		if out.call != nil && (out.call.PrincipalID != current.CreatedBy || out.call.WorkspaceID != current.WorkspaceID || thread.ConversationID != "comment:"+out.call.RootCommentID) {
			return out, ErrAuthorizationForbidden
		}
		break
	}
	if out.call == nil {
		return out, nil
	}
	out.evidence, err = s.readCommentDeliveryEvidence(ctx, out.call.EvidenceEventID)
	if err != nil {
		return out, err
	}
	if !commentInvocationEvidenceMatches(out.evidence, current.TaskID, out.call) {
		return out, ErrAuthorizationForbidden
	}
	return out, nil
}
