package riidoaiserver

import (
	"errors"
	"maps"
)

func (s *Store) handlePollCommand(state *storeState, msg pollCmd) pollResult {
	if err := msg.ctx.Err(); err != nil {
		return pollResult{err: err}
	}
	_, authenticated := claimDeviceFromContext(msg.ctx)
	_, durable := s.operationStore.(AssignmentClaimer)
	if authenticated && !durable {
		if s.snapshotStore != nil && s.operationStore != nil {
			return pollResult{err: errors.New("authenticated claims require one atomic persistence authority")}
		}
		trial := cloneLegacyEventState(state)
		trial.executionClaims = maps.Clone(state.executionClaims)
		trial.devicePolicyRevisions = maps.Clone(state.devicePolicyRevisions)
		trial.tasks = maps.Clone(state.tasks)
		trial.agentAssignments = maps.Clone(state.agentAssignments)
		quiet := *s
		quiet.outbox = nil
		result := quiet.commitPollCommand(&trial, msg)
		if result.err != nil {
			result.response = PollResponse{}
			return result
		}
		events := eventsAfterSeq(&trial, state.nextEventSeq)
		trial.subscribers = state.subscribers
		*state = trial
		for _, event := range events {
			s.publishTaskEvent(state, event)
		}
		return result
	}
	// Durable claimers commit the claim transaction before actor publication.
	result := s.commitPollCommand(state, msg)
	if authenticated && result.err != nil {
		result.response = PollResponse{}
	}
	return result
}

func (s *Store) commitPollCommand(state *storeState, msg pollCmd) pollResult {
	beforeEventSeq := state.nextEventSeq
	response, saved, mutated, kind, err := s.handlePoll(msg.ctx, state, msg.agentID, msg.req, msg.countRequest)
	if err == nil && mutated != nil {
		err = s.saveOperation(state, kind, *mutated, eventsAfterSeq(state, beforeEventSeq))
		if err == nil {
			err = s.saveSnapshot(state)
		}
	}
	if err == nil && mutated == nil && response.Action == PollStart {
		if !saved {
			err = s.saveAssignmentMutationOperations(state, AssignmentOperationPollStart, *response.Assignment, eventsAfterSeq(state, beforeEventSeq))
		}
		if err == nil {
			err = s.saveSnapshot(state)
		}
	}
	if err == nil {
		for _, event := range eventsAfterSeq(state, beforeEventSeq) {
			logAssignmentFailure(state, event)
		}
	}
	return pollResult{response: response, operationAlreadySaved: saved, mutatedAssignment: mutated, mutationOperationType: kind, err: err}
}
