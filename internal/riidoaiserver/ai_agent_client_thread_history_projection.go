package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentTaskThreadHistory(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadHistoryCollectionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadHistoryCollectionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return AIAgentTaskThreadHistoryCollectionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads, snapshots := s.visibleTaskThreadHistoryRecordsLocked(principal, taskID)
	response := AIAgentTaskThreadHistoryCollectionResponse{
		SchemaVersion:  SchemaVersion,
		TaskID:         taskID,
		Threads:        threads,
		AgentSnapshots: snapshots,
	}
	suppressSupersededQueuedHistoryMessages(response.Threads)
	response.ActiveStream = taskThreadHistoryActiveStream(response.Threads)
	return response, nil
}

func (s *DevelopmentAIAgentClientStore) taskThreadHistoryRecordLocked(principal AuthorizationResult, thread AIAgentTaskThreadRecord) AIAgentTaskThreadHistoryRecord {
	snapshotID := taskThreadAgentSnapshotID(thread.AgentSnapshot)
	record := AIAgentTaskThreadHistoryRecord{
		ThreadID:        thread.ThreadID,
		ConversationID:  taskThreadConversationID(thread),
		ParentThreadID:  taskThreadParentThreadID(thread),
		TaskID:          thread.TaskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         thread.AgentID,
		AgentSnapshotID: snapshotID,
		RunID:           thread.RunID,
		WorkStatus:      thread.WorkStatus,
		AssignmentState: thread.AssignmentState,
		StartedAt:       thread.StartedAt,
		CompletedAt:     thread.CompletedAt,
		Messages:        s.taskThreadHistoryMessagesLocked(thread),
	}
	if taskThreadHasActiveStream(thread) {
		link := activeStreamLinkForThread(thread, strings.TrimSpace(principal.WorkspaceID))
		record.ActiveStream = &link
	}
	return record
}

func (s *DevelopmentAIAgentClientStore) taskThreadHistoryMessagesLocked(thread AIAgentTaskThreadRecord) []AIAgentTaskThreadHistoryMessage {
	progress := s.cachedTaskThreadProgressMessagesLocked(thread)
	out := buildTaskThreadHistoryMessages(
		s.taskThreadMessages[thread.ThreadID],
		progress,
	)
	if message, ok := taskThreadProjectionMessage(thread); ok {
		out = append(out, message)
	}
	sortTaskThreadHistoryMessages(out)
	return out
}
