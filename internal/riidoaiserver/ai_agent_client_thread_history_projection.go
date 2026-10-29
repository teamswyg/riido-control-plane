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
	ids := s.visibleTaskMemberIDsLocked(principal, taskID)
	threads, snapshots := s.visibleTaskThreadHistoryRecordsLocked(principal, taskID)
	response := AIAgentTaskThreadHistoryCollectionResponse{
		SchemaVersion:    SchemaVersion,
		AssignedAgentIDs: ids,
		TaskID:           taskID,
		Threads:          threads,
		AgentSnapshots:   snapshots,
	}
	s.projectAgentMessageReactionsLocked(response.Threads)
	suppressSupersededQueuedHistoryMessages(response.Threads)
	response.ActiveStream = taskThreadHistoryActiveStream(response.Threads)
	return response, nil
}

func (s *DevelopmentAIAgentClientStore) taskThreadHistoryRecordLocked(thread AIAgentTaskThreadRecord, streamHref string) AIAgentTaskThreadHistoryRecord {
	record := AIAgentTaskThreadHistoryRecord{
		ExecutionContext:   copyExecutionContext(thread.ExecutionContext),
		FailureDiagnostics: clientVisibleFailureDiagnostics(thread.FailureDiagnostics),
		ThreadID:           thread.ThreadID,
		ConversationID:     taskThreadConversationID(thread),
		ParentThreadID:     taskThreadParentThreadID(thread),
		TaskID:             thread.TaskID,
		AssignmentID:       thread.AssignmentID,
		AgentID:            thread.AgentID,
		AgentSnapshotID:    thread.AgentSnapshotID,
		RunID:              thread.RunID,
		WorkStatus:         thread.WorkStatus,
		AssignmentState:    thread.AssignmentState,
		StartedAt:          thread.StartedAt,
		CompletedAt:        thread.CompletedAt,
		Messages:           s.taskThreadHistoryMessagesLocked(thread),
	}
	if taskThreadHasActiveStream(thread) {
		link := activeStreamLinkForThreadHref(thread, streamHref)
		record.ActiveStream = &link
	}
	return record
}

func (s *DevelopmentAIAgentClientStore) taskThreadHistoryMessagesLocked(thread AIAgentTaskThreadRecord) []AIAgentTaskThreadHistoryMessage {
	stored := s.taskThreadMessages[thread.ThreadID]
	if len(stored) == 0 {
		return s.cachedTaskThreadHistoryMessagesLocked(thread)
	}
	progress := s.cachedTaskThreadProgressMessagesLocked(thread)
	out := buildTaskThreadHistoryMessages(
		stored,
		progress,
	)
	if message, ok := taskThreadProjectionMessage(thread); ok {
		found := false
		for i := range out {
			if out[i].MessageID == message.MessageID {
				message.Reactions = out[i].Reactions
				out[i] = message
				found = true
				break
			}
		}
		if !found {
			out = append(out, message)
		}
	}
	sortTaskThreadHistoryMessages(out)
	return out
}
