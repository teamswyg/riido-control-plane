package riidoaiserver

import "context"

func (s *Store) HeartbeatAgent(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error) {
	response, _, err := s.HeartbeatAgentWithEvents(ctx, agentID, req)
	return response, err
}

func (s *Store) HeartbeatAgentWithEvents(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []TaskEvent, error) {
	reply := make(chan heartbeatResult, 1)
	if err := s.send(ctx, heartbeatCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentHeartbeatResponse{}, nil, err
	}
	select {
	case res := <-reply:
		if res.err != nil {
			return AgentHeartbeatResponse{}, nil, res.err
		}
		var events []TaskEvent
		for _, mutation := range res.mutations {
			events = append(events, mutation.events...)
		}
		return res.response, events, nil
	case <-ctx.Done():
		return AgentHeartbeatResponse{}, nil, ctx.Err()
	}
}
