package riidoaiserver

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/teamswyg/riido-contracts/reactionemoji"
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
	// revision orders member reaction notifications within one instance.
	revision uint64
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
