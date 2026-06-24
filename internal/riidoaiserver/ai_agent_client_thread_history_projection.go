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
	threads := s.visibleTaskThreadsLocked(principal, taskID)
	response := AIAgentTaskThreadHistoryCollectionResponse{
		SchemaVersion:  SchemaVersion,
		TaskID:         taskID,
		Threads:        make([]AIAgentTaskThreadHistoryRecord, 0, len(threads)),
		AgentSnapshots: map[string]AIAgentTaskThreadAgentSnapshot{},
	}
	for _, thread := range threads {
		record := s.taskThreadHistoryRecordLocked(principal, thread)
		response.Threads = append(response.Threads, record)
		if record.AgentSnapshotID != "" && thread.AgentSnapshot != nil {
			response.AgentSnapshots[record.AgentSnapshotID] = *copyTaskThreadAgentSnapshot(thread.AgentSnapshot)
		}
	}
	response.ActiveStream = taskThreadHistoryActiveStream(response.Threads)
	if len(response.AgentSnapshots) == 0 {
		response.AgentSnapshots = nil
	}
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
	out := copyTaskThreadHistoryMessages(s.taskThreadMessages[thread.ThreadID])
	out = append(out, taskThreadProgressMessages(thread)...)
	if message, ok := taskThreadProjectionMessage(thread); ok {
		out = append(out, message)
	}
	sortTaskThreadHistoryMessages(out)
	return out
}
