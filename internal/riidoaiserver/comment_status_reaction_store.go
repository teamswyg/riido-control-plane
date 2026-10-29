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
	return s.updateTaskThreadReactionState(ctx, workspaceID, threadID, func(state *AIAgentTaskThreadReactionState) bool {
		if state.CommentEventID != "" || strings.TrimSpace(eventID) == "" {
			return false
		}
		state.CommentEventID = strings.TrimSpace(eventID)
		return true
	})
}

func (s *DevelopmentAIAgentClientStore) MarkCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) error {
	return s.updateTaskThreadReactionState(ctx, workspaceID, threadID, func(state *AIAgentTaskThreadReactionState) bool {
		if state.CommentEventID == "" || state.CommentReactionApplied == emoji {
			return false
		}
		state.CommentReactionApplied = emoji
		return true
	})
}

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
		if emoji := agentStatusReactionEmoji(thread.AssignmentState); emoji != state.CommentReactionApplied {
			plans = append(plans, commentStatusReactionPlan{WorkspaceID: s.taskThreadWorkspaceIDLocked(thread), ThreadID: thread.ThreadID, EventID: state.CommentEventID, Emoji: emoji})
		}
	}
	return plans
}

func (s *DevelopmentAIAgentClientStore) updateTaskThreadReactionState(ctx context.Context, workspaceID, threadID string, update func(*AIAgentTaskThreadReactionState) bool) error {
	thread, err := s.FindAIAgentTaskThreadByID(ctx, workspaceID, threadID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.taskThreadReactions[thread.ThreadID]
	if update(&state) {
		s.setTaskThreadReactionStateLocked(thread.ThreadID, state)
	}
	return nil
}
