package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

func (s *Store) applyClaimedAssignment(state *storeState, claim AssignmentClaimResult) error {
	if !claim.Claimed {
		return nil
	}
	if claim.Assignment.ID == "" {
		return errors.New("claimed assignment_id is required")
	}
	if claim.Operation.OperationID == "" {
		return errors.New("claimed assignment operation_id is required")
	}
	if err := validateAssignmentOperationRecord(claim.Operation); err != nil {
		return err
	}
	if claim.Operation.OperationType != AssignmentOperationPollStart {
		return fmt.Errorf("claimed assignment operation_type = %q", claim.Operation.OperationType)
	}
	if claim.Operation.AssignmentID != claim.Assignment.ID {
		return fmt.Errorf("claimed assignment operation assignment_id mismatch %q", claim.Operation.AssignmentID)
	}
	if claim.Operation.Assignment != claim.Assignment {
		return errors.New("claimed assignment operation assignment mismatch")
	}
	return s.applyClaimedAssignmentOperations(state, claim)
}

func handleLoadAssignmentProjection(state *storeState, assignmentID string) (AssignmentProjection, bool, error) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("assignment_id is required")
	}
	assignment, ok := state.assignments[assignmentID]
	if !ok {
		return AssignmentProjection{}, false, nil
	}
	lastEventSeq := int64(0)
	for _, event := range state.events[assignment.TaskID] {
		if event.AssignmentID == assignmentID && event.Seq > lastEventSeq {
			lastEventSeq = event.Seq
		}
	}
	return AssignmentProjection{Assignment: assignment, LastEventSeq: lastEventSeq}, true, nil
}
