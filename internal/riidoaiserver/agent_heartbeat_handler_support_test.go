package riidoaiserver

import "context"

type handlerHeartbeatEventStore struct {
	*handlerAssignmentStore
	heartbeatEvents func(context.Context, string, AgentHeartbeatRequest) (AgentHeartbeatResponse, []TaskEvent, error)
}

func (s *handlerHeartbeatEventStore) HeartbeatAgentWithEvents(
	ctx context.Context,
	agentID string,
	req AgentHeartbeatRequest,
) (AgentHeartbeatResponse, []TaskEvent, error) {
	return s.heartbeatEvents(ctx, agentID, req)
}
