package riidoaiserver

import "context"

// RecordAIAgentContextReads appends one agent_context_read event to the
// thread of the read assignment. Reads for an unknown thread are not streamed;
// they remain in the assignment read records.
func (s *DevelopmentAIAgentClientStore) RecordAIAgentContextReads(ctx context.Context, reads []AssignmentContextRead) error {
	if err := ctx.Err(); err != nil || len(reads) == 0 {
		return err
	}
	event := newAgentContextReadEvent(reads)
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.taskThreadForAssignmentLocked(event.TaskID, event.AgentID, event.AssignmentID)
	if !ok {
		return nil
	}
	event.ThreadID, event.RunID = thread.ThreadID, thread.RunID
	s.appendClientEventLocked(event.EventType, event)
	return nil
}
