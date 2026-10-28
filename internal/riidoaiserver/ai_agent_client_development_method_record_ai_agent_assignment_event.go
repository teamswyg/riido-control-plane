package riidoaiserver

import (
	"context"

	"github.com/teamswyg/riido-contracts/assignment"
)

func (s *DevelopmentAIAgentClientStore) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	input, err := newAssignmentEventInput(agentID, req, event)
	if err != nil {
		return err
	}

	if assignment.RecoveryContextFromMetadata(input.Metadata) != nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if message, found, err := decodeNativeHistory(input.Metadata); err != nil || found {
		if err != nil {
			return err
		}
		return s.recordNativeHistoryLocked(input, message)
	}
	agent, ok := s.agents[input.AgentID]
	if !ok {
		return ErrAIAgentNotFound
	}
	thread, hadThread := s.assignmentEventThreadLocked(input)
	previousThread := thread
	if hadThread && !agentAssignmentStateAcceptsRuntimeProgress(previousThread.AssignmentState) {
		// Snapshot repair can observe the durable terminal event before its fanout.
		// Deliver that event without reviving or mutating the terminal projection.
		if thread.AssignmentID == input.AssignmentID && isTerminal(input.State) && input.Type != EventRiidoLog && input.Type != EventProviderLog && input.Type != EventProviderError {
			response := assignmentEventActionResponse(thread, input.State, input.Message, input.Metadata)
			if response.WorkStatus == thread.WorkStatus {
				s.appendTaskNotificationLocked(input.Event, response)
			}
		}
		return nil
	}

	if input.IsProgressLog() {
		return s.recordAssignmentProgressEventLocked(input, agent, hadThread, previousThread, thread)
	}
	s.recordAssignmentStatusEventLocked(input, agent, hadThread, previousThread, thread)
	return nil
}
