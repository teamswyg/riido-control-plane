package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DevelopmentAIAgentClientStore) MutateTaskMessageReaction(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, remove bool) (AIAgentTaskThreadMessageReactionsResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadMessageReactionsResponse{}, err
	}
	emoji, err := normalizeMessageReactionEmoji(emoji)
	if err != nil {
		return AIAgentTaskThreadMessageReactionsResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.visibleTaskThreadLocked(principal, taskID, threadID)
	if !ok {
		return AIAgentTaskThreadMessageReactionsResponse{}, ErrAIAgentNotFound
	}
	messages := s.taskThreadHistoryMessagesLocked(thread)
	for _, message := range messages {
		if message.MessageID != messageID {
			continue
		}
		if (message.Role != AIAgentTaskThreadMessageRoleUser && message.Role != AIAgentTaskThreadMessageRoleAgent) || message.CommentKind == AgentTaskCommentRuntimeProgress || message.CommentKind == AgentTaskCommentAssignmentStarted || message.CommentKind == AgentTaskCommentQueuedByBusyAgent {
			return AIAgentTaskThreadMessageReactionsResponse{}, errors.New("progress messages are not reactable")
		}
		message.Reactions = mutateMessageReaction(message.Reactions, AIAgentMessageReactionActor{Kind: "member", ID: principal.PrincipalID}, emoji, remove)
		source := s.taskThreadMessages[threadID]
		replaced := false
		for i := range source {
			if source[i].MessageID == messageID {
				source[i].Reactions = message.Reactions
				replaced = true
				break
			}
		}
		if !replaced {
			source = append(source, message)
		}
		s.taskThreadMessages[threadID] = source
		s.dropTaskThreadHistoryCacheLocked(threadID)
		records, _ := s.visibleTaskThreadHistoryRecordsLocked(principal, taskID)
		s.projectAgentMessageReactionsLocked(records)
		for _, record := range records {
			if record.ThreadID == threadID {
				for _, current := range record.Messages {
					if current.MessageID == messageID {
						message.Reactions = current.Reactions
					}
				}
			}
		}
		s.messageReactionRevision++
		return AIAgentTaskThreadMessageReactionsResponse{SchemaVersion: SchemaVersion, TaskID: taskID, ThreadID: threadID, MessageID: messageID, Reactions: message.Reactions, revision: s.messageReactionRevision}, nil
	}
	return AIAgentTaskThreadMessageReactionsResponse{}, ErrAIAgentNotFound
}
