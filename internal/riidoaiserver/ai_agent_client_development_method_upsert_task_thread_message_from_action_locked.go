package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) upsertTaskThreadMessageFromActionLocked(response AIAgentTaskActionResponse, sourceMessageID string) {
	now := time.Now().UTC()
	thread := AIAgentTaskThreadRecord{
		ThreadID:           response.ThreadID,
		TaskID:             response.TaskID,
		AssignmentID:       response.AssignmentID,
		AgentID:            response.AgentID,
		AgentSnapshot:      copyTaskThreadAgentSnapshot(response.AgentSnapshot),
		RunID:              response.RunID,
		SourceMessageID:    strings.TrimSpace(sourceMessageID),
		WorkStatus:         response.WorkStatus,
		AssignmentState:    response.AssignmentState,
		CommentKind:        response.CommentKind,
		Message:            response.Message,
		ResultMessage:      response.ResultMessage,
		FailureDiagnostics: copyFailureDiagnostics(response.FailureDiagnostics),
		StartedAt:          now,
		Lines:              []AgentThreadProgressLine{},
	}
	if !taskThreadHasActiveStream(thread) {
		thread.CompletedAt = now
	}
	s.ensureTaskThreadAgentSnapshotLocked(&thread, now)
	threads := s.taskThreads[response.TaskID]
	for i := range threads {
		if threads[i].ThreadID != response.ThreadID {
			continue
		}
		threads[i].RunID = response.RunID
		if strings.TrimSpace(response.AssignmentID) != "" {
			threads[i].AssignmentID = strings.TrimSpace(response.AssignmentID)
		}
		if threads[i].AgentSnapshot == nil {
			threads[i].AgentSnapshot = copyTaskThreadAgentSnapshot(response.AgentSnapshot)
			s.ensureTaskThreadAgentSnapshotLocked(&threads[i], now)
		}
		threads[i].WorkStatus = response.WorkStatus
		threads[i].AssignmentState = response.AssignmentState
		threads[i].FailureDiagnostics = copyFailureDiagnostics(response.FailureDiagnostics)
		threads[i].CommentKind = response.CommentKind
		threads[i].Message = response.Message
		threads[i].ResultMessage = response.ResultMessage
		if sourceMessageID != "" {
			threads[i].SourceMessageID = sourceMessageID
		}
		if threads[i].StartedAt.IsZero() {
			threads[i].StartedAt = now
		}
		if taskThreadHasActiveStream(threads[i]) {
			threads[i].CompletedAt = time.Time{}
		} else {
			threads[i].CompletedAt = now
		}
		s.taskThreads[response.TaskID] = threads
		return
	}
	s.taskThreads[response.TaskID] = append(threads, thread)
}
