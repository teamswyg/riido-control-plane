package riidoaiserver

import "context"

func (s *DevelopmentAIAgentClientStore) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	input, err := newAssignmentEventInput(agentID, req, event)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[input.AgentID]
	if !ok {
		return ErrAIAgentNotFound
	}
	thread, hadThread := s.assignmentEventThreadLocked(input)
	previousThread := thread

	if input.IsProgressLog() {
		return s.recordAssignmentProgressEventLocked(input, agent, hadThread, previousThread, thread)
	}
	s.recordAssignmentStatusEventLocked(input, agent, hadThread, previousThread, thread)
	return nil
}
