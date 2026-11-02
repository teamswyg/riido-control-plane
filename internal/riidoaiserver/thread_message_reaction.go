package riidoaiserver

import (
	"context"
	"errors"
	"github.com/teamswyg/riido-contracts/reactionemoji"
	"regexp"
	"slices"
	"strings"
)

type AIAgentMessageReactionActor struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}
type AIAgentMessageReaction struct {
	Emoji  string                        `json:"emoji"`
	Count  int                           `json:"count"`
	Actors []AIAgentMessageReactionActor `json:"actors"`
}
type AIAgentTaskThreadMessageReactionsResponse struct {
	SchemaVersion string                   `json:"schema_version"`
	TaskID        string                   `json:"task_id"`
	ThreadID      string                   `json:"thread_id"`
	MessageID     string                   `json:"message_id"`
	Reactions     []AIAgentMessageReaction `json:"reactions"`
}
type taskMessageReactionStore interface {
	MutateTaskMessageReaction(context.Context, AuthorizationResult, string, string, string, string, bool) (AIAgentTaskThreadMessageReactionsResponse, error)
}

var reactionEmojiPattern = regexp.MustCompile(`^[0-9a-f]{2,6}(-[0-9a-f]{2,6})*$`)

func normalizeMessageReactionEmoji(emoji string) (string, error) {
	emoji = strings.ToLower(strings.TrimSpace(emoji))
	if len(emoji) > 64 || !reactionEmojiPattern.MatchString(emoji) {
		return "", errors.New("emoji must use normalized unified codepoints")
	}
	return reactionemoji.Normalize(emoji), nil
}

func copyMessageReactions(in []AIAgentMessageReaction) []AIAgentMessageReaction {
	out := make([]AIAgentMessageReaction, len(in))
	for i, reaction := range in {
		out[i] = reaction
		out[i].Actors = slices.Clone(reaction.Actors)
	}
	return out
}

func mutateMessageReaction(in []AIAgentMessageReaction, actor AIAgentMessageReactionActor, emoji string, remove bool) []AIAgentMessageReaction {
	out := copyMessageReactions(in)
	for i := range out {
		if out[i].Emoji != emoji {
			continue
		}
		found := slices.Contains(out[i].Actors, actor)
		if remove {
			out[i].Actors = slices.DeleteFunc(out[i].Actors, func(a AIAgentMessageReactionActor) bool { return a == actor })
		} else if !found {
			out[i].Actors = append(out[i].Actors, actor)
		}
		out[i].Count = len(out[i].Actors)
		if out[i].Count == 0 {
			out = slices.Delete(out, i, i+1)
		}
		return out
	}
	if !remove {
		out = append(out, AIAgentMessageReaction{Emoji: emoji, Count: 1, Actors: []AIAgentMessageReactionActor{actor}})
	}
	return out
}

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
		return AIAgentTaskThreadMessageReactionsResponse{SchemaVersion: SchemaVersion, TaskID: taskID, ThreadID: threadID, MessageID: messageID, Reactions: message.Reactions}, nil
	}
	return AIAgentTaskThreadMessageReactionsResponse{}, ErrAIAgentNotFound
}

func (s *PersistentAIAgentClientStore) MutateTaskMessageReaction(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, remove bool) (AIAgentTaskThreadMessageReactionsResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadMessageReactionsResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.MutateTaskMessageReaction(ctx, principal, taskID, threadID, messageID, emoji, remove)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
