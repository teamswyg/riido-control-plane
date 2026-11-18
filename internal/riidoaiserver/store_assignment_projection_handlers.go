package riidoaiserver

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/teamswyg/riido-contracts/assignment"
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
	if !reflect.DeepEqual(claim.Operation.Assignment, claim.Assignment) {
		return errors.New("claimed assignment operation assignment mismatch")
	}
	return s.applyClaimedAssignmentOperations(state, claim)
}

func handleLoadAssignmentProjection(state *storeState, assignmentID string) (AssignmentProjection, bool, error) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AssignmentProjection{}, false, errors.New("assignment_id is required")
	}
	current, ok := state.assignments[assignmentID]
	if !ok {
		return AssignmentProjection{}, false, nil
	}
	projection := AssignmentProjection{Assignment: *copyAssignment(current)}
	terminalSeq := int64(0)
	var recoveryEvents []TaskEvent
	for _, event := range state.events[current.TaskID] {
		if event.AssignmentID != assignmentID {
			continue
		}
		if current.State == AssignmentFailed && event.Type == EventRiidoLog && event.Metadata[assignment.RecoveryStateKey] != "" && len(recoveryEvents) <= maxRecoveryReceipts {
			recoveryEvents = append(recoveryEvents, event)
		}
		if isTerminal(event.State) && (terminalSeq == 0 || event.Seq < terminalSeq) {
			terminalSeq, projection.ContextTerminalAt = event.Seq, event.At
		}
		if event.Seq > projection.LastEventSeq {
			projection.LastEventSeq = event.Seq
			projection.WorkspacePreparationFailed = isWorkspacePreparationFailure(current, event)
			projection.TerminalEvent = nil
			if assignmentTerminalEventMatches(current, event) {
				event.Metadata = cloneMetadata(event.Metadata)
				projection.TerminalEvent = &event
			}
		}
	}
	// Oversized legacy journals cannot establish preflight failure.
	projection.Recovery, _ = recoveryWatermarkForRecord(AssignmentOperationRecord{Events: recoveryEvents})
	return projection, true, nil
}
