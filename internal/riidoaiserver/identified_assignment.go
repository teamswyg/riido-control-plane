package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// Stable IDs let retrying publishers and restarted API instances converge on one request.
func identifiedAssignmentID(taskID string, req AssignRequest) string {
	return fmt.Sprintf("asn-request-%x", sha256.Sum256([]byte(req.WorkspaceID+"\x00"+taskID+"\x00"+req.AgentID+"\x00"+req.CreatedBy+"\x00"+req.RequestID)))
}

func (s *Store) assignIdentifiedRequest(state *storeState, taskID string, req AssignRequest) (Assignment, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	taskID = strings.TrimSpace(taskID)
	if len(req.RequestID) > 256 || req.WorkspaceID == "" || req.CreatedBy == "" {
		return Assignment{}, errors.New("request_id requires workspace and author scope (maximum 256 bytes)")
	}
	assignmentID := identifiedAssignmentID(taskID, req)
	if reader, ok := s.operationStore.(AssignmentProjectionReader); ok {
		projection, found, err := reader.LoadAssignmentProjection(context.Background(), assignmentID)
		if err != nil {
			return Assignment{}, err
		}
		if found {
			applyAssignmentToState(state, projection.Assignment)
			return projection.Assignment, nil
		}
	}
	if existing, ok := state.assignments[assignmentID]; ok {
		return existing, nil
	}
	// Do not publish an accepted event or allow polling until persistence succeeds.
	trial, err := stateFromSnapshot(snapshotFromState(state, s.now()))
	if err != nil {
		return Assignment{}, err
	}
	quiet := *s
	quiet.outbox = nil
	beforeSeq := trial.nextEventSeq
	assignment, err := quiet.handleAssign(&trial, taskID, req, true, false)
	if err != nil {
		return Assignment{}, err
	}
	events := eventsAfterSeq(&trial, beforeSeq)
	if err = s.saveAssignmentMutationOperations(&trial, AssignmentOperationAssignTask, assignment, events); err != nil {
		return Assignment{}, err
	}
	if reader, ok := s.operationStore.(AssignmentProjectionReader); ok {
		projection, found, readErr := reader.LoadAssignmentProjection(context.Background(), assignmentID)
		if readErr != nil {
			return Assignment{}, readErr
		}
		if found {
			assignment = projection.Assignment
			applyAssignmentToState(&trial, assignment)
		}
	}
	if err = s.saveSnapshot(&trial); err != nil {
		return Assignment{}, err
	}
	state.assignments, state.agentAssignments, state.tasks, state.events = trial.assignments, trial.agentAssignments, trial.tasks, trial.events
	state.nextAssignmentSeq, state.nextEventSeq = trial.nextAssignmentSeq, trial.nextEventSeq
	for _, event := range events {
		if s.outbox != nil {
			s.appendTaskEventToOutbox(state, event)
		}
		for _, ch := range state.subscribers[event.TaskID] {
			select {
			case ch <- event:
			default:
			}
		}
	}
	s.signalAgentWaiters(state, assignment.AgentID)
	return assignment, nil
}
