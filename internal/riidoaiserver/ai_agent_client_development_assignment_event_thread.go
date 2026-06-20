package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) assignmentEventThreadLocked(input assignmentEventInput) (AIAgentTaskThreadRecord, bool) {
	thread, ok := s.taskThreadForAssignmentLocked(input.TaskID, input.AgentID, input.AssignmentID)
	if !ok {
		thread, ok = s.activeTaskThreadForAgentLocked(input.TaskID, input.AgentID)
	}
	if ok {
		return thread, true
	}
	runID := "run-" + input.AssignmentID
	thread = AIAgentTaskThreadRecord{
		ThreadID:        threadIDForRun(input.TaskID, input.AgentID, runID),
		TaskID:          input.TaskID,
		AssignmentID:    input.AssignmentID,
		AgentID:         input.AgentID,
		RunID:           runID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         assignmentEventVisibleThreadMessage(input.State, input.Type, input.Message, ""),
		StartedAt:       input.At,
		Lines:           []AgentThreadProgressLine{},
	}
	if thread.StartedAt.IsZero() {
		thread.StartedAt = time.Now().UTC()
	}
	s.taskThreads[input.TaskID] = append(s.taskThreads[input.TaskID], thread)
	return thread, false
}
