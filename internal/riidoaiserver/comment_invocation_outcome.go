package riidoaiserver

import (
	"context"
	"time"
)

type commentInvocationRecorder interface {
	RecordCommentInvocationOutcome(context.Context, string, string, string, string, string) error
}

func (s *DevelopmentAIAgentClientStore) RecordCommentInvocationOutcome(ctx context.Context, taskID, threadID, key, outcome, notice string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.taskThreads[taskID] {
		thread := &s.taskThreads[taskID][i]
		if thread.ThreadID != threadID || thread.CommentInvocation == nil {
			continue
		}
		binding := copyCommentInvocation(thread.CommentInvocation)
		if binding.Outcomes[key] != "" {
			return nil
		}
		if binding.Outcomes == nil {
			binding.Outcomes = map[string]string{}
		}
		binding.Outcomes[key] = outcome
		thread.CommentInvocation = binding
		if notice != "" {
			s.appendThreadHistoryMessageLocked(threadID, AIAgentTaskThreadHistoryMessage{MessageID: commentInvocationRequestID(thread.AssignmentID, key), Role: "system", AssignmentID: thread.AssignmentID, Body: notice, ObservedAt: time.Now().UTC()})
			s.appendAgentTaskActionEvent(actionResponseFromThread(*thread, binding.WorkspaceID))
		}
		return nil
	}
	return ErrAIAgentTaskThreadConflict
}

func (s *PersistentAIAgentClientStore) RecordCommentInvocationOutcome(ctx context.Context, taskID, threadID, key, outcome, notice string) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	if err := s.DevelopmentAIAgentClientStore.RecordCommentInvocationOutcome(ctx, taskID, threadID, key, outcome, notice); err != nil {
		return err
	}
	return s.saveSnapshot(ctx)
}
