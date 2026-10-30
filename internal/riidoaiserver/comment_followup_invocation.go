package riidoaiserver

import (
	"context"
	"strings"
)

// A fresh human request starts a new bounded call chain, but its authority still
// comes from the original, server-verified comment. Never accept this binding
// from the HTTP request or transfer another requester's evidence.
func (s Server) prepareCommentFollowup(ctx context.Context, principal AuthorizationResult, selected AIAgentTaskThreadRecord, target string) (*CommentInvocation, error) {
	if !strings.HasPrefix(selected.ConversationID, "comment:") {
		return nil, nil
	}
	collection, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, selected.TaskID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]AIAgentTaskThreadRecord, len(collection.Threads))
	for _, row := range collection.Threads {
		byID[row.ThreadID] = row
	}
	// Older followups dropped the binding. Only follow persisted parent edges
	// within this task/root; a matching conversation string alone is insufficient.
	seen := make(map[string]bool)
	id := selected.ThreadID
	for id != "" && !seen[id] {
		seen[id] = true
		row, ok := byID[id]
		if !ok || row.TaskID != selected.TaskID || row.ConversationID != selected.ConversationID {
			return nil, ErrAuthorizationForbidden
		}
		if call := row.CommentInvocation; call != nil {
			if call.PrincipalID != principal.PrincipalID || call.WorkspaceID != principal.WorkspaceID {
				return nil, nil
			}
			if row.ConversationID != "comment:"+call.RootCommentID {
				return nil, ErrAuthorizationForbidden
			}
			evidence, err := s.readCommentDeliveryEvidence(ctx, call.EvidenceEventID)
			if err != nil {
				return nil, err
			}
			if !commentInvocationEvidenceMatches(evidence, selected.TaskID, call) {
				return nil, ErrAuthorizationForbidden
			}
			next := newCommentInvocation(evidence)
			next.EvidenceAgentID = evidence.AgentID
			next.Ancestors = []string{target}
			next.HumanFollowup = true
			return next, nil
		}
		id = row.ParentThreadID
	}
	return nil, nil
}
