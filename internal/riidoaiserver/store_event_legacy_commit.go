package riidoaiserver

import (
	"context"
	"errors"
	"maps"
	"slices"
)

// Legacy local snapshots and write-only journals have no cross-store atomic
// receipt port. Publish and retain local ACKs only after one authority commits.
func (s *Store) commitLegacyAgentEvent(ctx context.Context, state *storeState, agentID string, req AgentEventRequest, approval *recoveryCandidateApproval) (AgentEventResponse, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return AgentEventResponse{}, assignmentDependencyFailure(errors.New("legacy dual storage requires atomic projection receipts"))
	}
	trial := cloneLegacyEventState(state)
	quiet := *s
	quiet.outbox = nil
	quiet.recoveryApproval = approval
	response, err := quiet.handleEvent(&trial, agentID, req)
	if err != nil {
		return AgentEventResponse{}, err
	}
	events := eventsAfterSeq(&trial, state.nextEventSeq)
	if len(events) == 0 {
		return response, nil
	}
	if err := s.persistLegacyAgentEvent(ctx, &trial, *response.Assignment, events); err != nil {
		return AgentEventResponse{}, assignmentDependencyFailure(err)
	}
	trial.subscribers = state.subscribers
	*state = trial
	for _, event := range events {
		s.publishTaskEvent(state, event)
		observeCommittedAssignmentEvent(ctx, state, event)
	}
	return response, nil
}

func (s *Store) persistLegacyAgentEvent(ctx context.Context, trial *storeState, current Assignment, events []TaskEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.operationStore != nil {
		return s.operationStore.SaveAssignmentOperation(ctx, AssignmentOperationRecord{
			SchemaVersion: AssignmentOperationSchemaVersion,
			OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, current, events), OperationType: AssignmentOperationAgentEvent,
			TaskID: current.TaskID, AssignmentID: current.ID, AgentID: current.AgentID, Assignment: current, Events: events, RecordedAt: s.now(),
		})
	}
	if s.snapshotStore != nil {
		return s.snapshotStore.SaveStoreSnapshot(ctx, snapshotFromState(trial, s.now()))
	}
	return nil
}

func cloneLegacyEventState(state *storeState) storeState {
	trial := *state
	trial.assignments = maps.Clone(state.assignments)
	trial.events = maps.Clone(state.events)
	for taskID, events := range trial.events {
		trial.events[taskID] = slices.Clone(events)
	}
	trial.assignmentEventACKs = maps.Clone(state.assignmentEventACKs)
	for id, receipts := range trial.assignmentEventACKs {
		trial.assignmentEventACKs[id] = slices.Clone(receipts)
	}
	trial.subscribers = nil
	return trial
}
