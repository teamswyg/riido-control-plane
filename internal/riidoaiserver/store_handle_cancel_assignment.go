package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Store) handleCancelAssignment(state *storeState, taskID string, req CancelAssignmentRequest) (Assignment, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.Reason = strings.TrimSpace(req.Reason)
	if taskID == "" {
		return Assignment{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return Assignment{}, errors.New("agent_id is required")
	}
	if req.durableProjection != nil {
		projection := *req.durableProjection
		if reader, ok := s.operationStore.(AssignmentProjectionReader); ok {
			current, found, err := reader.LoadAssignmentProjection(context.Background(), req.AssignmentID)
			if err != nil {
				return Assignment{}, err
			}
			if !found {
				return Assignment{}, errors.New("membership cancellation projection disappeared")
			}
			projection = current
		}
		if projection.Assignment.ID != req.AssignmentID || projection.Assignment.TaskID != taskID || projection.Assignment.AgentID != req.AgentID {
			return Assignment{}, errors.New("membership cancellation projection scope mismatch")
		}
		applyAssignmentProjectionToState(state, projection)
	}
	assignment, ok := state.assignmentForClientStop(taskID, req)
	if !ok {
		if req.AssignmentID != "" {
			return Assignment{}, fmt.Errorf("assignment %s not found", req.AssignmentID)
		}
		return Assignment{}, fmt.Errorf("active assignment for task %s and agent %s not found", taskID, req.AgentID)
	}
	if assignment.TaskID != taskID {
		return Assignment{}, fmt.Errorf("assignment %s belongs to task %s", assignment.ID, assignment.TaskID)
	}
	if assignment.AgentID != req.AgentID {
		return Assignment{}, fmt.Errorf("assignment %s belongs to agent %s", assignment.ID, assignment.AgentID)
	}
	if isTerminal(assignment.State) || assignment.State.Code() == AssignmentStateCodeCancelling {
		return assignment, nil
	}

	nextState := AssignmentCancelling
	eventType := EventAssignmentCancelling
	message := req.Reason
	if message == "" {
		message = "assignment cancellation requested by client"
	}
	if assignment.State.Code() == AssignmentStateCodeQueued {
		nextState = AssignmentCancelled
		eventType = EventAssignmentCancelled
		if req.Reason == "" {
			message = "queued assignment was cancelled by client"
		}
	}
	if !canTransitionAssignment(assignment.State, nextState) {
		return Assignment{}, fmt.Errorf("invalid assignment transition %s -> %s", assignment.State, nextState)
	}
	now := s.now()
	assignment.State = nextState
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, eventType, assignment.State, message, map[string]string{"source": "client_stop"}, now)
	// Wake a daemon parked on a long-poll so it observes the cancel (PollCancel)
	// within the hold instead of waiting for the budget to elapse.
	s.signalAgentWaiters(state, assignment.AgentID)
	return assignment, nil
}
