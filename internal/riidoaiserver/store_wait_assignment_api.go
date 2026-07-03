package riidoaiserver

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// WaitForAssignment is the long-poll claim. The wait loop runs on the caller's
// goroutine so the actor keeps servicing assign/heartbeat/poll commands.
func (s *Store) WaitForAssignment(ctx context.Context, agentID string, req PollRequest, hold, tick time.Duration) (response PollResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationWaitAssignment)
	startedAt := time.Now()
	waited := false
	defer func() {
		span.SetAttributes(
			BoolTraceAttribute(riidoPollWaitedKey, waited),
			Int64TraceAttribute(riidoPollElapsedMsKey, pollDurationMs(time.Since(startedAt))),
			Int64TraceAttribute(riidoPollHoldMsKey, pollDurationMs(hold)),
			Int64TraceAttribute(riidoPollTickMsKey, pollDurationMs(tick)),
		)
		if err == nil {
			span.SetAttributes(StringTraceAttribute(metadatakeys.RiidoPollAction.String(), string(response.Action)))
		}
		s.observeStoreOperation(StoreOperationWaitAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	resp, err := s.PollAgent(ctx, agentID, req)
	if err != nil || resp.Action != PollNone || hold <= 0 {
		return resp, err
	}
	signal, release, err := s.registerWaiter(ctx, agentID)
	if err != nil {
		return PollResponse{}, err
	}
	defer release()
	resp, err = s.pollAgent(ctx, agentID, req, false)
	if err != nil || resp.Action != PollNone {
		return resp, err
	}
	waited = true
	if s.operationStore == nil {
		return s.waitForAssignmentSignalOnly(ctx, agentID, req, signal, hold)
	}
	return s.waitForAssignmentSignal(ctx, agentID, req, signal, hold, tick)
}

func (s *Store) waitForAssignmentSignal(ctx context.Context, agentID string, req PollRequest, signal <-chan struct{}, hold, tick time.Duration) (PollResponse, error) {
	if tick <= 0 || tick > hold {
		tick = hold
	}
	deadline := time.NewTimer(hold)
	defer deadline.Stop()
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-signal:
			if resp, err := s.pollAgent(ctx, agentID, req, false); err != nil || resp.Action != PollNone {
				return resp, err
			}
		case <-ticker.C:
			if resp, err := s.pollAgent(ctx, agentID, req, false); err != nil || resp.Action != PollNone {
				return resp, err
			}
		case <-deadline.C:
			return PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}, nil
		case <-ctx.Done():
			return PollResponse{}, ctx.Err()
		}
	}
}
