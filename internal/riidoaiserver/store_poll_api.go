package riidoaiserver

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) PollAgent(ctx context.Context, agentID string, req PollRequest) (PollResponse, error) {
	return s.pollAgent(ctx, agentID, req, true)
}

// pollAgent runs a single point-in-time evaluation. count records whether this
// evaluation should be tallied as a daemon poll request.
func (s *Store) pollAgent(ctx context.Context, agentID string, req PollRequest, count bool) (response PollResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationPollAssignment)
	startedAt := time.Now()
	defer func() {
		if err == nil {
			span.SetAttributes(StringTraceAttribute(metadatakeys.RiidoPollAction.String(), string(response.Action)))
		}
		if !count {
			FinishTraceSpan(span, err)
			return
		}
		s.observeStoreOperation(StoreOperationPollAssignment, startedAt, err)
		if err == nil && response.Action == PollStart {
			s.observeStoreOperation(StoreOperationLeaseAssignment, startedAt, nil)
		}
		FinishTraceSpan(span, err)
	}()
	reply := make(chan pollResult, 1)
	if err := s.send(ctx, pollCmd{agentID: agentID, req: req, countRequest: count, reply: reply}); err != nil {
		return PollResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return PollResponse{}, ctx.Err()
	}
}
