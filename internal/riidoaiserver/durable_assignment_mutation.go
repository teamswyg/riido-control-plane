package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
)

var errAssignmentProjectionConflict = errors.New("assignment changed concurrently; retry the request")

// Cancellation and daemon events share a conditional operation+projection commit.
// Nothing is broadcast before the durable write has won the race.
func (s *Store) mutateDurableAssignment(ctx context.Context, state *storeState, id string, kind AssignmentOperationType, mutate func(*Store, *storeState) (Assignment, error), observe ...func(TaskEvent)) (Assignment, error) {
	reader := s.operationStore.(AssignmentProjectionReader)
	for range 3 {
		if err := ctx.Err(); err != nil {
			return Assignment{}, err
		}
		expected, found, err := reader.LoadAssignmentProjection(ctx, id)
		if err != nil {
			return Assignment{}, assignmentDependencyFailure(err)
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
		quiet.recoveryFence = expected.Recovery
		quiet.eventACKFence = expected.EventACKs
		quiet.eventACK = nil
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
			RecordedAt: s.now(), ExpectedProjection: &expected, eventACK: quiet.eventACK,
		}
		if _, err := encodeAssignmentEventACKs(record); err != nil {
			return Assignment{}, err
		}
		if err := ctx.Err(); err != nil {
			return Assignment{}, err
		}
		err = s.operationStore.SaveAssignmentOperation(ctx, record)
		if errors.Is(err, errAssignmentProjectionConflict) {
			continue
		}
		if err != nil {
			return Assignment{}, assignmentDependencyFailure(err)
		}
		if err = s.saveSnapshot(&trial); err != nil {
			return Assignment{}, assignmentDependencyFailure(err)
		}
		applyAssignmentToState(state, assignment)
		state.agentEventsTotal = trial.agentEventsTotal
		for _, event := range events {
			s.appendRecordedEvent(state, event)
			for _, observer := range observe {
				observer(event)
			}
		}
		s.signalAgentWaiters(state, assignment.AgentID)
		return assignment, nil
	}
	return Assignment{}, errAssignmentProjectionConflict
}
