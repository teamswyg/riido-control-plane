package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Store) handlePoll(ctx context.Context, state *storeState, agentID string, req PollRequest, count bool) (PollResponse, bool, *Assignment, AssignmentOperationType, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return PollResponse{}, false, nil, "", errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, req); err != nil {
		if response, ok := deletedAgentPollCancelResponse(state, agentID, req, count); ok {
			return response, false, nil, "", nil
		}
		return PollResponse{}, false, nil, "", err
	}
	if count {
		state.pollRequestsTotal++
	}
	response := PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}
	assignIDs := state.agentAssignments[agentID]
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State.Code() == AssignmentStateCodeCancelling {
			expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
			if err != nil {
				return PollResponse{}, false, nil, "", err
			}
			if expired {
				stale := s.failStaleAssignment(state, assignment)
				recordPollAction(state, response.Action, count)
				return response, false, &stale, AssignmentOperationAgentEvent, nil
			}
			response.Action = PollCancel
			response.Assignment = copyAssignment(assignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	durableAssignment, staleAssignment, ok, err := s.loadDurableActiveAssignment(state, agentID, s.now())
	if err != nil {
		return PollResponse{}, false, nil, "", err
	}
	if staleAssignment != nil {
		recordPollAction(state, response.Action, count)
		return response, false, staleAssignment, AssignmentOperationAgentEvent, nil
	}
	if ok {
		switch {
		case durableAssignment.State.Code() == AssignmentStateCodeCancelling:
			response.Action = PollCancel
			response.Assignment = copyAssignment(durableAssignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		case isAgentActive(durableAssignment.State):
			response.Action = PollActive
			response.Assignment = copyAssignment(durableAssignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if isAgentActive(assignment.State) {
			expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
			if err != nil {
				return PollResponse{}, false, nil, "", err
			}
			if expired {
				stale := s.failStaleAssignment(state, assignment)
				recordPollAction(state, response.Action, count)
				return response, false, &stale, AssignmentOperationAgentEvent, nil
			}
			response.Action = PollActive
			response.Assignment = copyAssignment(assignment)
			recordPollAction(state, response.Action, count)
			return response, false, nil, "", nil
		}
	}
	if claimer, ok := s.operationStore.(AssignmentClaimer); ok {
		claim, err := claimer.ClaimNextAssignment(ctx, agentID, s.now())
		if err != nil {
			return PollResponse{}, false, nil, "", err
		}
		if claim.Claimed {
			if expected := newExecutionClaim(ctx, claim.Assignment); expected != nil && (claim.Operation.ExecutionClaim == nil || *claim.Operation.ExecutionClaim != *expected) {
				return PollResponse{}, false, nil, "", errors.New("authenticated claim persistence evidence is missing")
			}
			if err := s.applyClaimedAssignment(state, claim); err != nil {
				return PollResponse{}, false, nil, "", err
			}
			response.Action = PollStart
			response.Assignment = copyAssignment(claim.Assignment)
			recordPollAction(state, response.Action, count)
			return response, true, nil, "", nil
		}
		recordPollAction(state, response.Action, count)
		return response, false, nil, "", nil
	}
	for _, id := range assignIDs {
		assignment := state.assignments[id]
		if assignment.State.Code() != AssignmentStateCodeQueued {
			continue
		}
		if err := s.repairQueuedAssignmentBlockerForClaim(state, &assignment); err != nil {
			return PollResponse{}, false, nil, "", err
		}
		if !assignmentBlockerCleared(state, assignment) {
			continue
		}
		now := s.now()
		assignment.BlockedByAssignmentID = ""
		assignment.State = AssignmentLeased
		assignment.LeaseToken = fmt.Sprintf("%s:%d", assignment.ID, now.UnixNano())
		assignment.UpdatedAt = now
		if claim := newExecutionClaim(ctx, assignment); claim != nil {
			if err := rememberExecutionClaim(state, claim); err != nil {
				return PollResponse{}, false, nil, "", err
			}
		}
		state.assignments[id] = assignment
		s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentLeased, assignment.State, "", map[string]string{"lease_token": assignment.LeaseToken}, now)
		response.Action = PollStart
		response.Assignment = copyAssignment(assignment)
		recordPollAction(state, response.Action, count)
		return response, false, nil, "", nil
	}
	recordPollAction(state, response.Action, count)
	return response, false, nil, "", nil
}
