package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
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
	req = assignmentEventACKRequest(assignment, req)
	if err := s.validateReadyEventLease(state, assignment, &req); err != nil {
		return AgentEventResponse{}, err
	}
	if response, duplicate, err := s.checkStateAssignmentEventACK(state, assignment, req); err != nil || duplicate {
		return response, err
	}
	var recoveryErr error
	req, recoveryErr = s.validateRecoveryEvent(state, assignment, req)
	if recoveryErr != nil {
		return AgentEventResponse{}, recoveryErr
	}
	if existing, duplicate, err := s.checkRecoveryWatermark(assignment, req); err != nil || duplicate {
		return AgentEventResponse{SchemaVersion: SchemaVersion, Assignment: copyAssignment(assignment), Event: existing}, err
	}
	if existing, ok := duplicateAssignmentEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, strings.TrimSpace(req.EventType), req.Metadata); ok {
		if incoming := assignmentcontract.RecoveryContextFromMetadata(req.Metadata); incoming != nil {
			previous := assignmentcontract.RecoveryContextFromMetadata(existing.Metadata)
			if previous == nil || *incoming != *previous {
				return AgentEventResponse{}, errors.New("recovery event key was reused with another transition")
			}
		}
		return AgentEventResponse{SchemaVersion: SchemaVersion, Assignment: copyAssignment(assignment), Event: existing}, nil
	}
	if req.EventType == assignmentcontract.EventAssignmentWaitingForUser && (!req.ConfirmationReason.Valid() || isTerminal(assignment.State) || assignment.State == AssignmentCancelling) {
		return AgentEventResponse{}, errors.New("waiting event requires an active execution and confirmation_reason")
	}
	if event, ok := repeatedNotificationEvent(state, assignment, req); ok {
		return AgentEventResponse{SchemaVersion: SchemaVersion, Assignment: copyAssignment(assignment), Event: event}, nil
	}
	payloadHash, err := assignmentEventACKPayloadHash(assignment, req)
	if err != nil {
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
	eventType := req.EventType
	if eventType == assignmentcontract.EventAssignmentWaitingForUser {
		assignment.ConfirmationReason = req.ConfirmationReason
	} else if eventType == EventAssignmentRunning || isTerminal(assignment.State) {
		assignment.ConfirmationReason = ""
	}
	state.assignments[assignment.ID] = assignment
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
	s.retainStateAssignmentEventACK(state, assignment, req, event, payloadHash)
	return AgentEventResponse{
		SchemaVersion: SchemaVersion,
		Assignment:    copyAssignment(assignment),
		Event:         event,
	}, nil
}
