package riidoaiserver

import (
	"strings"
	"time"
)

// agentThreadReactionNotification is the monolith request that leaves an
// unread Riido notification for a member reaction (RIID-5906).
type agentThreadReactionNotification struct {
	WorkspaceID          string `json:"workspace_id"`
	TaskID               string `json:"task_id"`
	ThreadID             string `json:"thread_id"`
	MessageID            string `json:"message_id"`
	RecipientPrincipalID string `json:"recipient_principal_id"`
	ActorPrincipalID     string `json:"actor_principal_id"`
	Emoji                string `json:"emoji"`
	Excerpt              string `json:"excerpt"`
	// revision orders notifications of the same message and actor.
	revision uint64
}

// toggledMemberReactions returns a new slice; the stored slice is shared with
// snapshots and is never modified in place.
func toggledMemberReactions(members []AIAgentTaskThreadMemberReaction, reaction AIAgentTaskThreadMemberReaction, add bool) ([]AIAgentTaskThreadMemberReaction, bool) {
	out := make([]AIAgentTaskThreadMemberReaction, 0, len(members)+1)
	found := false
	for _, existing := range members {
		same := existing.Target == reaction.Target && existing.Emoji == reaction.Emoji && existing.PrincipalID == reaction.PrincipalID
		if same {
			found = true
			if !add {
				continue
			}
		}
		out = append(out, existing)
	}
	if add && !found {
		reaction.CreatedAt = time.Now().UTC()
		return append(out, reaction), true
	}
	return out, !add && found
}

func (s *DevelopmentAIAgentClientStore) setTaskThreadReactionStateLocked(threadID string, state AIAgentTaskThreadReactionState) {
	next := make(map[string]AIAgentTaskThreadReactionState, len(s.taskThreadReactions)+1)
	for id, existing := range s.taskThreadReactions {
		next[id] = existing
	}
	next[threadID] = state
	s.taskThreadReactions = next
}

// memberReactionNotification asks monolith to notify the author of a user
// message about another member's reaction. The latest remaining emoji of the
// actor is sent, and "" removes the notification. Agent messages are skipped.
func memberReactionNotification(principal AuthorizationResult, thread AIAgentTaskThreadRecord, message AIAgentTaskThreadHistoryMessage, members []AIAgentTaskThreadMemberReaction, target string) *agentThreadReactionNotification {
	actor := strings.TrimSpace(principal.PrincipalID)
	author := strings.TrimSpace(message.AuthorPrincipalID)
	if message.Role != AIAgentTaskThreadMessageRoleUser || author == "" || author == actor {
		return nil
	}
	emoji := ""
	for _, reaction := range members {
		if reaction.Target == target && reaction.PrincipalID == actor {
			emoji = reaction.Emoji
		}
	}
	return &agentThreadReactionNotification{
		WorkspaceID: principal.WorkspaceID, TaskID: thread.TaskID, ThreadID: thread.ThreadID,
		MessageID: message.MessageID, RecipientPrincipalID: author, ActorPrincipalID: actor,
		Emoji: emoji, Excerpt: message.Body,
	}
}
