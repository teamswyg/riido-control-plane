package riidoaiserver

import (
	"context"
	"strings"
)

// commentStatusReactionPlan is a monolith comment whose agent status reaction
// differs from the reaction last written there.
type commentStatusReactionPlan struct {
	WorkspaceID string
	ThreadID    string
	EventID     string
	Emoji       string
}

type commentStatusReactionStore interface {
	RecordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) error
	CommentStatusReactionPlans(taskID string) []commentStatusReactionPlan
	MarkCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) error
}

// RecordCommentDeliveryEvent links a run thread to the monolith delivery that
// started it. The event id is server-side evidence, never a client value.
func (s *DevelopmentAIAgentClientStore) RecordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) error {
	_, err := s.recordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID)
	return err
}

func (s *DevelopmentAIAgentClientStore) recordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) (bool, error) {
	return s.updateTaskThreadReactionState(ctx, workspaceID, threadID, func(state *AIAgentTaskThreadReactionState) bool {
		if state.CommentEventID != "" || strings.TrimSpace(eventID) == "" {
			return false
		}
		state.CommentEventID = strings.TrimSpace(eventID)
		return true
	})
}

func (s *DevelopmentAIAgentClientStore) MarkCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) error {
	_, err := s.markCommentStatusReactionApplied(ctx, workspaceID, threadID, emoji)
	return err
}

func (s *DevelopmentAIAgentClientStore) markCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) (bool, error) {
	return s.updateTaskThreadReactionState(ctx, workspaceID, threadID, func(state *AIAgentTaskThreadReactionState) bool {
		if state.CommentEventID == "" || state.CommentReactionApplied == emoji {
			return false
		}
		state.CommentReactionApplied = emoji
		return true
	})
}

func (s *DevelopmentAIAgentClientStore) updateTaskThreadReactionState(ctx context.Context, workspaceID, threadID string, update func(*AIAgentTaskThreadReactionState) bool) (bool, error) {
	thread, err := s.FindAIAgentTaskThreadByID(ctx, workspaceID, threadID)
	if err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.taskThreadReactions[thread.ThreadID]
	changed := update(&state)
	if changed {
		s.setTaskThreadReactionStateLocked(thread.ThreadID, state)
	}
	return changed, nil
}
