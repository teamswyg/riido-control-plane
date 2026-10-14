package riidoaiserver

import (
	"context"
	"fmt"
	"strings"
)

func (s *Store) saveSnapshot(state *storeState) error {
	if s.snapshotStore == nil {
		return nil
	}
	return s.snapshotStore.SaveStoreSnapshot(context.Background(), snapshotFromState(state, s.now()))
}

func (s *Store) saveOperation(state *storeState, operationType AssignmentOperationType, assignment Assignment, events []TaskEvent) error {
	if s.operationStore == nil || len(events) == 0 {
		return nil
	}
	recordedAt := s.now()
	return s.operationStore.SaveAssignmentOperation(context.Background(), AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(operationType, assignment, events),
		OperationType: operationType,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        append([]TaskEvent(nil), events...),
		RecordedAt:    recordedAt,
	})
}

func (s *Store) saveAssignmentMutationOperations(state *storeState, primaryType AssignmentOperationType, primary Assignment, events []TaskEvent) error {
	if len(events) == 0 {
		return nil
	}
	eventsByAssignment := map[string][]TaskEvent{}
	var assignmentIDs []string
	for _, event := range events {
		assignmentID := strings.TrimSpace(event.AssignmentID)
		if assignmentID == "" {
			continue
		}
		if _, ok := eventsByAssignment[assignmentID]; !ok {
			assignmentIDs = append(assignmentIDs, assignmentID)
		}
		eventsByAssignment[assignmentID] = append(eventsByAssignment[assignmentID], event)
	}
	for _, assignmentID := range assignmentIDs {
		assignment := state.assignments[assignmentID]
		if assignment.ID == "" {
			return fmt.Errorf("assignment %s not found for mutation events", assignmentID)
		}
		operationType := AssignmentOperationAgentEvent
		if assignmentID == primary.ID {
			operationType = primaryType
			assignment = primary
		}
		if err := s.saveOperation(state, operationType, assignment, eventsByAssignment[assignmentID]); err != nil {
			return err
		}
	}
	return nil
}
