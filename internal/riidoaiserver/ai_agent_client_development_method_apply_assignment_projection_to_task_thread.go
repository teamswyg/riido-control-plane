package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) applyAssignmentProjectionToTaskThread(thread AIAgentTaskThreadRecord, projection AssignmentProjection, diagnostics *AIAgentTaskThreadQueueDiagnostics) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	threads := s.taskThreads[thread.TaskID]
	for i := range threads {
		if threads[i].ThreadID != thread.ThreadID ||
			strings.TrimSpace(threads[i].AssignmentID) != strings.TrimSpace(thread.AssignmentID) ||
			!taskThreadHasActiveStream(threads[i]) {
			continue
		}
		previous := threads[i]
		s.ensureTaskThreadAgentSnapshotLocked(&threads[i], threads[i].StartedAt)
		response := assignmentEventActionResponse(threads[i], projection.Assignment.State, "", nil)
		response.AssignmentID = projection.Assignment.ID
		if previous.AssignmentID == response.AssignmentID &&
			previous.WorkStatus == response.WorkStatus &&
			previous.AssignmentState == response.AssignmentState &&
			previous.CommentKind == response.CommentKind &&
			queueDiagnosticsEqual(previous.QueueDiagnostics, diagnostics) {
			return false
		}
		threads[i].AssignmentID = response.AssignmentID
		threads[i].WorkStatus = response.WorkStatus
		threads[i].AssignmentState = response.AssignmentState
		threads[i].QueueDiagnostics = copyQueueDiagnostics(diagnostics)
		threads[i].FailureDiagnostics = copyFailureDiagnostics(response.FailureDiagnostics)
		threads[i].CommentKind = response.CommentKind
		threads[i].Message = response.Message
		threads[i].ResultMessage = response.ResultMessage
		if assignmentStateIsTerminal(projection.Assignment.State) {
			completedAt := projection.Assignment.UpdatedAt
			if completedAt.IsZero() {
				completedAt = time.Now().UTC()
			} else {
				completedAt = completedAt.UTC()
			}
			threads[i].CompletedAt = completedAt
		} else {
			threads[i].CompletedAt = time.Time{}
		}
		s.taskThreads[thread.TaskID] = threads

		agent := s.agents[response.AgentID]
		if agent.AgentID != "" {
			if taskThreadHasActiveStream(AIAgentTaskThreadRecord{AssignmentState: response.AssignmentState}) {
				if agent.AssignedTaskCount == 0 {
					agent.AssignedTaskCount = 1
				}
			} else if agent.AssignedTaskCount > 0 {
				agent.AssignedTaskCount--
			}
			agent.WorkStatus = response.WorkStatus
			agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
			if taskThreadHasActiveStream(AIAgentTaskThreadRecord{AssignmentState: response.AssignmentState}) {
				agent.Editability = AgentEditabilityBlockedAssignedTasks
			}
			s.agents[agent.AgentID] = agent
		}
		if shouldFanoutAgentTaskActionEvent(true, previous, response) {
			s.appendAgentTaskActionEvent(response)
		}
		return true
	}
	return false
}
