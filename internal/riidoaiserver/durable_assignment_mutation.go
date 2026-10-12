package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
)

var errAssignmentProjectionConflict = errors.New("assignment changed concurrently; retry the request")

// Cancellation and daemon events share a conditional operation+projection commit.
// Nothing is broadcast before the durable write has won the race.
func (s *Store) mutateDurableAssignment(state *storeState, id string, kind AssignmentOperationType, mutate func(*Store, *storeState) (Assignment, error)) (Assignment, error) {
	reader := s.operationStore.(AssignmentProjectionReader)
	for range 3 {
		expected, found, err := reader.LoadAssignmentProjection(context.Background(), id)
		if err != nil {
			return Assignment{}, err
		}
		if !found {
			return Assignment{}, fmt.Errorf("assignment %s not found", id)
		}
		trial, err := stateFromSnapshot(snapshotFromState(state, s.now()))
		if err != nil {
			return Assignment{}, err
		}
		applyAssignmentProjectionToState(&trial, expected)
		before := trial.nextEventSeq
		quiet := *s
		quiet.outbox = nil
		assignment, err := mutate(&quiet, &trial)
		if err != nil {
			return Assignment{}, err
		}
		events := eventsAfterSeq(&trial, before)
		if len(events) == 0 {
			applyAssignmentProjectionToState(state, expected)
			return assignment, nil
		}
		record := AssignmentOperationRecord{
			SchemaVersion: AssignmentOperationSchemaVersion,
			OperationID:   assignmentOperationID(kind, assignment, events),
			OperationType: kind, TaskID: assignment.TaskID, AssignmentID: assignment.ID,
			AgentID: assignment.AgentID, Assignment: assignment, Events: events,
			RecordedAt: s.now(), ExpectedProjection: &expected,
		}
		err = s.operationStore.SaveAssignmentOperation(context.Background(), record)
		if errors.Is(err, errAssignmentProjectionConflict) {
			continue
		}
		if err != nil {
			return Assignment{}, err
		}
		if err = s.saveSnapshot(&trial); err != nil {
			return Assignment{}, err
		}
		applyAssignmentToState(state, assignment)
		state.agentEventsTotal = trial.agentEventsTotal
		for _, event := range events {
			s.appendRecordedEvent(state, event)
		}
		s.signalAgentWaiters(state, assignment.AgentID)
		return assignment, nil
	}
	return Assignment{}, errAssignmentProjectionConflict
}
