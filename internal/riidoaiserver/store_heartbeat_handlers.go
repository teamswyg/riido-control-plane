package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Store) handleHeartbeat(state *storeState, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []heartbeatMutation, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentHeartbeatResponse{}, nil, errors.New("agent_id is required")
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return AgentHeartbeatResponse{}, nil, err
	}
	assignmentIDs := heartbeatAssignmentIDs(state, agentID, req)
	response := AgentHeartbeatResponse{SchemaVersion: SchemaVersion}
	if len(assignmentIDs) == 0 {
		return response, nil, nil
	}
	now := s.now()
	var mutations []heartbeatMutation
	leaseStore, _ := s.operationStore.(AssignmentActiveLeaseStore)
	for _, assignmentID := range assignmentIDs {
		assignment, ok := state.assignments[assignmentID]
		if !ok {
			return AgentHeartbeatResponse{}, nil, fmt.Errorf("assignment %s not found", assignmentID)
		}
		if assignment.AgentID != agentID {
			return AgentHeartbeatResponse{}, nil, fmt.Errorf("assignment %s belongs to agent %s", assignmentID, assignment.AgentID)
		}
		if !assignmentHoldsActiveLease(assignment.State) {
			continue
		}
		expired, err := s.assignmentActiveLeaseExpired(state, assignment, now)
		if err != nil {
			return AgentHeartbeatResponse{}, nil, err
		}
		if expired {
			beforeEventSeq := state.nextEventSeq
			stale := s.failStaleAssignment(state, assignment)
			mutations = append(mutations, heartbeatMutation{
				assignment:    stale,
				operationType: AssignmentOperationAgentEvent,
				events:        eventsAfterSeq(state, beforeEventSeq),
			})
			continue
		}
		if leaseStore != nil {
			if err := leaseStore.RefreshAgentActiveAssignment(context.Background(), assignment, now); err != nil {
				return AgentHeartbeatResponse{}, nil, err
			}
		}
		assignment.UpdatedAt = now
		state.assignments[assignment.ID] = assignment
		response.RefreshedAssignments = append(response.RefreshedAssignments, assignment)
	}
	return response, mutations, nil
}
