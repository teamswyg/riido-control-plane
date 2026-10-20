package riidoaiserver

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

func (s *PersistentAIAgentClientStore) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) (err error) {
	timing := newAgentProjectionTiming(req, event)
	defer func() { timing.finish(ctx, err != nil) }()
	// Recovery is already durable in the assignment journal. Reuse the no-op's
	// validation/cancellation path without accessing a snapshot or operation lock.
	if assignment.RecoveryContextFromMetadata(event.Metadata) != nil {
		return s.DevelopmentAIAgentClientStore.RecordAIAgentAssignmentEvent(ctx, agentID, req, event)
	}
	s.operationMu.Lock()
	timing.lock = time.Since(timing.started).Milliseconds()
	defer s.operationMu.Unlock()
	if err = measureProjectionPhase(&timing.reload, func() error { return s.reloadSnapshot(ctx) }); err != nil {
		return err
	}
	if err = measureProjectionPhase(&timing.projection, func() error {
		return s.DevelopmentAIAgentClientStore.RecordAIAgentAssignmentEvent(ctx, agentID, req, event)
	}); err != nil {
		return err
	}
	return measureProjectionPhase(&timing.save, func() error { return s.saveSnapshot(ctx) })
}
