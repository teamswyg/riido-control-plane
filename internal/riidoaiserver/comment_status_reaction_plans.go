package riidoaiserver

import "strings"

// CommentStatusReactionPlans reads in-memory state only; it is called from
// internal hooks that already hold a verified task id.
func (s *DevelopmentAIAgentClientStore) CommentStatusReactionPlans(taskID string) []commentStatusReactionPlan {
	s.mu.Lock()
	defer s.mu.Unlock()
	var plans []commentStatusReactionPlan
	for _, thread := range s.taskThreads[strings.TrimSpace(taskID)] {
		state, ok := s.taskThreadReactions[thread.ThreadID]
		if !ok || state.CommentEventID == "" {
			continue
		}
		if emoji := agentStatusReactionEmoji(thread.AssignmentState); appliedCommentReaction(emoji) != state.CommentReactionApplied {
			plans = append(plans, commentStatusReactionPlan{WorkspaceID: s.taskThreadWorkspaceIDLocked(thread), ThreadID: thread.ThreadID, EventID: state.CommentEventID, Emoji: emoji})
		}
	}
	return plans
}

// commentReactionCleared records a confirmed removal. An empty applied value
// only means nothing was confirmed: a 👀 write may have reached monolith
// before its record failed, so a run without a reaction still writes the
// removal once.
const commentReactionCleared = "-"

func appliedCommentReaction(emoji string) string {
	if emoji == "" {
		return commentReactionCleared
	}
	return emoji
}
