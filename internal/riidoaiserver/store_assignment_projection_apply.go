package riidoaiserver

import "errors"

func (s *Store) applyClaimedAssignmentOperations(state *storeState, claim AssignmentClaimResult) error {
	operations := claim.Operations
	if len(operations) == 0 {
		operations = []AssignmentOperationRecord{claim.Operation}
	}
	sawPrimary := false
	for _, operation := range operations {
		if err := validateAssignmentOperationRecord(operation); err != nil {
			return err
		}
		if operation.OperationID == claim.Operation.OperationID {
			sawPrimary = true
		}
	}
	if !sawPrimary {
		return errors.New("claimed assignment operations missing primary claim operation")
	}
	for _, operation := range operations {
		applyAssignmentToState(state, operation.Assignment)
		for _, event := range operation.Events {
			s.appendRecordedEvent(state, event)
		}
	}
	return nil
}

func applyAssignmentToState(state *storeState, assignment Assignment) {
	state.assignments[assignment.ID] = assignment
	if seq := assignmentSequence(assignment.ID); seq > state.nextAssignmentSeq {
		state.nextAssignmentSeq = seq
	}
	if !assignmentIDInAgentQueue(state.agentAssignments[assignment.AgentID], assignment.ID) {
		state.agentAssignments[assignment.AgentID] = append(state.agentAssignments[assignment.AgentID], assignment.ID)
	}
	task := state.tasks[assignment.TaskID]
	current := state.assignments[task.currentAssignmentID]
	if task.id == "" || assignmentIsNewer(assignment, current) || task.currentAssignmentID == assignment.ID {
		state.tasks[assignment.TaskID] = taskRecord{
			id:                  assignment.TaskID,
			componentID:         assignment.ComponentID,
			currentAssignmentID: assignment.ID,
		}
	}
}

func applyAssignmentProjectionToState(state *storeState, projection AssignmentProjection) {
	applyAssignmentToState(state, projection.Assignment)
	if projection.LastEventSeq > state.nextEventSeq {
		state.nextEventSeq = projection.LastEventSeq
	}
}
