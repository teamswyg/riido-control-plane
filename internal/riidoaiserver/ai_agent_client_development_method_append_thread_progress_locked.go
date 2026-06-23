package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) appendThreadProgressLocked(event AgentThreadProgressEvent) {
	now := time.Now().UTC()
	threads := s.taskThreads[event.TaskID]
	for i := range threads {
		if threads[i].ThreadID != event.ThreadID {
			continue
		}
		// Terminal/stop fence (C2), defense in depth: never let a runtime-progress
		// event re-activate a thread the user stopped or that already finished,
		// regardless of which ingestion path called us.
		if !agentAssignmentStateAcceptsRuntimeProgress(threads[i].AssignmentState) {
			return
		}
		threads[i].RunID = event.RunID
		if strings.TrimSpace(event.AssignmentID) != "" {
			threads[i].AssignmentID = strings.TrimSpace(event.AssignmentID)
		}
		s.ensureTaskThreadAgentSnapshotLocked(&threads[i], threads[i].StartedAt)
		threads[i].WorkStatus = event.WorkStatus
		threads[i].AssignmentState = event.AssignmentState
		threads[i].QueueDiagnostics = nil
		threads[i].CommentKind = event.CommentKind
		if len(event.Lines) > 0 {
			threads[i].Message = event.Lines[len(event.Lines)-1].Message
		}
		threads[i].Lines = mergeThreadProgressLines(threads[i].Lines, event.Lines)
		s.taskThreads[event.TaskID] = threads
		return
	}
	message := "agent progress updated"
	if len(event.Lines) > 0 {
		message = event.Lines[len(event.Lines)-1].Message
	}
	thread := AIAgentTaskThreadRecord{
		ThreadID:        event.ThreadID,
		TaskID:          event.TaskID,
		AssignmentID:    event.AssignmentID,
		AgentID:         event.AgentID,
		RunID:           event.RunID,
		WorkStatus:      event.WorkStatus,
		AssignmentState: event.AssignmentState,
		CommentKind:     event.CommentKind,
		Message:         message,
		StartedAt:       now,
		Lines:           copyProgressLines(event.Lines),
	}
	s.ensureTaskThreadAgentSnapshotLocked(&thread, now)
	s.taskThreads[event.TaskID] = append(threads, thread)
}
