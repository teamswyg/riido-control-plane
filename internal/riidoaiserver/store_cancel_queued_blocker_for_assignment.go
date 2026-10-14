package riidoaiserver

import (
	"strings"
	"time"
)

func (s *Store) cancelQueuedBlockerForAssignment(state *storeState, assignment *Assignment, now time.Time) bool {
	if assignment == nil || strings.TrimSpace(assignment.BlockedByAssignmentID) == "" {
		return false
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	if blocker.ID == "" || blocker.State.Code() != AssignmentStateCodeQueued {
		return false
	}
	blocker.State = AssignmentCancelled
	blocker.UpdatedAt = now
	state.assignments[blocker.ID] = blocker
	s.appendEvent(state, blocker.TaskID, blocker.ID, blocker.AgentID, EventAssignmentCancelled, blocker.State, "queued blocker was cancelled before daemon lease", nil, now)

	assignment.BlockedByAssignmentID = ""
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = *assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "queued blocker cleared before daemon lease", nil, now)
	s.signalAgentWaiters(state, assignment.AgentID)
	return true
}
