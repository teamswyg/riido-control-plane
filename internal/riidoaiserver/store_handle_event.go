package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

func (s *Store) handleEvent(state *storeState, agentID string, req AgentEventRequest) (AgentEventResponse, error) {
	agentID = strings.TrimSpace(agentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req = normalizeNeedsInputAssignmentEvent(req)
	if agentID == "" {
		return AgentEventResponse{}, errors.New("agent_id is required")
	}
	if req.AssignmentID == "" {
		return AgentEventResponse{}, errors.New("assignment_id is required")
	}
	assignment, ok := state.assignments[req.AssignmentID]
	if !ok {
		return AgentEventResponse{}, fmt.Errorf("assignment %s not found", req.AssignmentID)
	}
	if assignment.AgentID != agentID {
		return AgentEventResponse{}, fmt.Errorf("assignment %s belongs to agent %s", req.AssignmentID, assignment.AgentID)
	}
	if req.TaskID != "" && req.TaskID != assignment.TaskID {
		return AgentEventResponse{}, fmt.Errorf("assignment %s belongs to task %s", req.AssignmentID, assignment.TaskID)
	}
	if err := validateDaemonBinding(s.agentRegistry, agentID, PollRequest{DaemonID: req.DaemonID, DeviceID: req.DeviceID, RuntimeID: req.RuntimeID}); err != nil {
		return AgentEventResponse{}, err
	}
	state.agentEventsTotal++
	now := s.now()
	req.ProviderSessionID = strings.TrimSpace(req.ProviderSessionID)
	assignmentChanged := false
	if req.State != "" && req.State != assignment.State {
		if !canTransitionAssignment(assignment.State, req.State) {
			return AgentEventResponse{}, fmt.Errorf("invalid assignment transition %s -> %s", assignment.State, req.State)
		}
		assignment.State = req.State
		assignment.UpdatedAt = now
		assignmentChanged = true
	}
	if req.ProviderSessionID != "" && req.ProviderSessionID != assignment.ProviderSessionID {
		assignment.ProviderSessionID = req.ProviderSessionID
		assignment.UpdatedAt = now
		assignmentChanged = true
	}
	if assignmentChanged {
		state.assignments[assignment.ID] = assignment
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		if req.State != "" {
			eventType = EventAssignmentStateUpdated
		}
		if eventType == "" {
			eventType = EventRiidoLog
		}
	}
	if !assignmentChanged {
		if existing, ok := duplicateAssignmentEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, req.Metadata); ok {
			return AgentEventResponse{
				SchemaVersion: SchemaVersion,
				Assignment:    copyAssignment(assignment),
				Event:         existing,
			}, nil
		}
	}
	event := s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, assignment.State, req.Message, req.Metadata, now)
	return AgentEventResponse{
		SchemaVersion: SchemaVersion,
		Assignment:    copyAssignment(assignment),
		Event:         event,
	}, nil
}
