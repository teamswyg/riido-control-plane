package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) waitForAssignmentSignalOnly(
	ctx context.Context,
	agentID string,
	req PollRequest,
	signal <-chan struct{},
	hold time.Duration,
) (PollResponse, error) {
	deadline := time.NewTimer(hold)
	defer deadline.Stop()
	for {
		select {
		case <-signal:
			resp, err := s.pollAgent(ctx, agentID, req, false)
			if err != nil || resp.Action != PollNone {
				return resp, err
			}
		case <-deadline.C:
			return PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}, nil
		case <-ctx.Done():
			return PollResponse{}, ctx.Err()
		}
	}
}
