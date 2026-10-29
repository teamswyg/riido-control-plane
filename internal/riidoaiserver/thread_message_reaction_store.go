package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

const maxTaskThreadMemberReactions = 1000

var errThreadMessageReactionLimit = errors.New("task thread reaction limit reached")

type threadMessageReactionChange struct {
	Response AIAgentTaskThreadMessageReactionsResponse
	Changed  bool
	Notify   *agentThreadReactionNotification
}

// ReactToAIAgentTaskThreadMessage adds or deletes the caller's own member
// reaction. Repeating an add or deleting a missing reaction changes nothing.
func (s *DevelopmentAIAgentClientStore) ReactToAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, add bool) (threadMessageReactionChange, error) {
	if err := ctx.Err(); err != nil {
		return threadMessageReactionChange{}, err
	}
	emoji, err := normalizeReactionEmoji(emoji)
	if err != nil {
		return threadMessageReactionChange{}, err
	}
	principalID := strings.TrimSpace(principal.PrincipalID)
	if principalID == "" {
		return threadMessageReactionChange{}, errThreadMessageReactionMissing
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.visibleTaskThreadLocked(principal, strings.TrimSpace(taskID), strings.TrimSpace(threadID))
	if !ok {
		return threadMessageReactionChange{}, ErrAIAgentNotFound
	}
	messages := s.taskThreadHistoryMessagesLocked(thread)
	message, found := findTaskThreadHistoryMessage(messages, strings.TrimSpace(messageID))
	if !found {
		return threadMessageReactionChange{}, ErrAIAgentNotFound
	}
	target, ok := taskThreadReactionTarget(message)
	if !ok {
		return threadMessageReactionChange{}, errThreadMessageNotReactable
	}
	state := s.taskThreadReactions[thread.ThreadID]
	members, changed := toggledMemberReactions(state.Members, AIAgentTaskThreadMemberReaction{Target: target, Emoji: emoji, PrincipalID: principalID}, add)
	if changed && add && len(members) > maxTaskThreadMemberReactions {
		return threadMessageReactionChange{}, errThreadMessageReactionLimit
	}
	if changed {
		state.Members = members
		s.setTaskThreadReactionStateLocked(thread.ThreadID, state)
	}
	result := threadMessageReactionChange{Changed: changed, Response: AIAgentTaskThreadMessageReactionsResponse{ThreadID: thread.ThreadID, MessageID: message.MessageID, Reactions: []AIAgentMessageReaction{}}}
	if reacted, ok := findTaskThreadHistoryMessage(withTaskThreadReactions(thread, messages, state), message.MessageID); ok && reacted.Reactions != nil {
		result.Response.Reactions = reacted.Reactions
	}
	if changed {
		result.Notify = memberReactionNotification(principal, thread, message, members, target)
	}
	return result, nil
}

func findTaskThreadHistoryMessage(messages []AIAgentTaskThreadHistoryMessage, messageID string) (AIAgentTaskThreadHistoryMessage, bool) {
	for _, message := range messages {
		if messageID != "" && message.MessageID == messageID {
			return message, true
		}
	}
	return AIAgentTaskThreadHistoryMessage{}, false
}
