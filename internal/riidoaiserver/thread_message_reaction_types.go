package riidoaiserver

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// RIID-5907: emoji reactions on task thread messages (policy task_thread_message_reaction.v1).

type AIAgentMessageReactionActorKind string

const (
	AIAgentMessageReactionActorMember AIAgentMessageReactionActorKind = "member"
	AIAgentMessageReactionActorAgent  AIAgentMessageReactionActorKind = "agent"
)

type AIAgentMessageReactionActor struct {
	Kind AIAgentMessageReactionActorKind `json:"kind"`
	ID   string                          `json:"id"`
}

type AIAgentMessageReaction struct {
	Emoji  string                        `json:"emoji"`
	Count  int                           `json:"count"`
	Actors []AIAgentMessageReactionActor `json:"actors"`
}

type AddAIAgentTaskThreadMessageReactionRequest struct {
	Emoji string `json:"emoji"`
}

type AIAgentTaskThreadMessageReactionsResponse struct {
	ThreadID  string                   `json:"thread_id"`
	MessageID string                   `json:"message_id"`
	Reactions []AIAgentMessageReaction `json:"reactions"`
}

// AIAgentTaskThreadMemberReaction is a stored member reaction. Target is the
// message id of a user message, or the run key of an agent message, whose id
// changes with the run state.
type AIAgentTaskThreadMemberReaction struct {
	Target      string    `json:"target"`
	Emoji       string    `json:"emoji"`
	PrincipalID string    `json:"principal_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// AIAgentTaskThreadReactionState keeps the reactions of one task thread. For a
// run started by a monolith comment it also keeps the delivery event and the
// status reaction already written to that comment.
type AIAgentTaskThreadReactionState struct {
	Members                []AIAgentTaskThreadMemberReaction `json:"members,omitempty"`
	CommentEventID         string                            `json:"comment_event_id,omitempty"`
	CommentReactionApplied string                            `json:"comment_reaction_applied,omitempty"`
}

var (
	reactionEmojiPattern            = regexp.MustCompile(`^[0-9a-f]{2,6}(-[0-9a-f]{2,6})*$`)
	errThreadMessageReactionEmoji   = errors.New("emoji must be lowercase hex code points joined by '-' (max 64 bytes)")
	errThreadMessageNotReactable    = errors.New("progress messages are not reactable")
	errThreadMessageReactionMissing = errors.New("principal_id is required")
)

func normalizeReactionEmoji(emoji string) (string, error) {
	emoji = strings.TrimSpace(emoji)
	if len(emoji) > 64 || !reactionEmojiPattern.MatchString(emoji) {
		return "", errThreadMessageReactionEmoji
	}
	return emoji, nil
}
