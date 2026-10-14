package riidoaiserver

import "strings"

func (s *Store) repairQueuedAssignmentBlockerForClaim(state *storeState, assignment *Assignment) error {
	if assignment == nil || assignment.State.Code() != AssignmentStateCodeQueued || strings.TrimSpace(assignment.BlockedByAssignmentID) == "" {
		return nil
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	if blocker.ID == "" {
		blockedByID := assignment.BlockedByAssignmentID
		now := s.now()
		assignment.BlockedByAssignmentID = ""
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = *assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "missing blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": blockedByID}, now)
		s.signalAgentWaiters(state, assignment.AgentID)
		return nil
	}
	if isTerminal(blocker.State) {
		return nil
	}
	now := s.now()
	if blocker.State.Code() == AssignmentStateCodeQueued {
		blocker.State = AssignmentCancelled
		blocker.UpdatedAt = now
		state.assignments[blocker.ID] = blocker
		s.appendEvent(state, blocker.TaskID, blocker.ID, blocker.AgentID, EventAssignmentCancelled, blocker.State, "queued blocker was cancelled before queued assignment claim", map[string]string{"blocked_assignment_id": assignment.ID}, now)
		assignment.BlockedByAssignmentID = ""
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = *assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "queued blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": blocker.ID}, now)
		s.signalAgentWaiters(state, assignment.AgentID)
		return nil
	}
	if !assignmentHoldsActiveLease(blocker.State) {
		return nil
	}
	expired, err := s.assignmentActiveLeaseExpired(state, blocker, now)
	if err != nil {
		return err
	}
	if !expired {
		return nil
	}
	stale := s.failStaleAssignmentWithMessage(state, blocker, "blocked queued assignment repaired after stale blocker lease expired", map[string]string{
		"blocked_assignment_id": assignment.ID,
	})
	assignment.BlockedByAssignmentID = ""
	assignment.UpdatedAt = stale.UpdatedAt
	state.assignments[assignment.ID] = *assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentQueued, assignment.State, "stale blocker cleared before daemon lease", map[string]string{"blocked_by_assignment_id": stale.ID}, stale.UpdatedAt)
	s.signalAgentWaiters(state, assignment.AgentID)
	return nil
}
