package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

type handlerAssignmentStore struct {
	poll      func(context.Context, string, PollRequest) (PollResponse, error)
	heartbeat func(context.Context, string, AgentHeartbeatRequest) (AgentHeartbeatResponse, error)
	record    func(context.Context, string, AgentEventRequest) (AgentEventResponse, error)
}

func (s *handlerAssignmentStore) AssignTask(context.Context, string, AssignRequest) (Assignment, error) {
	return Assignment{}, errors.New("not implemented")
}

func (s *handlerAssignmentStore) AssignTaskReplacement(context.Context, string, AssignRequest) (Assignment, error) {
	return Assignment{}, errors.New("not implemented")
}

func (s *handlerAssignmentStore) AssignTaskAdditive(context.Context, string, AssignRequest) (Assignment, error) {
	return Assignment{}, errors.New("not implemented")
}

func (s *handlerAssignmentStore) PollAgent(ctx context.Context, agentID string, req PollRequest) (PollResponse, error) {
	if s.poll != nil {
		return s.poll(ctx, agentID, req)
	}
	return PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}, nil
}

func (s *handlerAssignmentStore) HeartbeatAgent(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error) {
	if s.heartbeat != nil {
		return s.heartbeat(ctx, agentID, req)
	}
	return AgentHeartbeatResponse{SchemaVersion: SchemaVersion}, nil
}

func (s *handlerAssignmentStore) RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (AgentEventResponse, error) {
	if s.record != nil {
		return s.record(ctx, agentID, req)
	}
	return AgentEventResponse{SchemaVersion: SchemaVersion, Event: TaskEvent{AgentID: agentID}}, nil
}

func (s *handlerAssignmentStore) SubscribeTask(context.Context, string) ([]TaskEvent, <-chan TaskEvent, func(), error) {
	return nil, nil, func() {}, nil
}

func (s *handlerAssignmentStore) Metrics(context.Context) (MetricsSnapshot, error) {
	return MetricsSnapshot{SchemaVersion: MetricsSchemaVersion}, nil
}

func (s *handlerAssignmentStore) Close() {}

type handlerLongPollStore struct {
	*handlerAssignmentStore
	wait func(context.Context, string, PollRequest, time.Duration, time.Duration) (PollResponse, error)
}

func (s *handlerLongPollStore) WaitForAssignment(ctx context.Context, agentID string, req PollRequest, hold, tick time.Duration) (PollResponse, error) {
	return s.wait(ctx, agentID, req, hold, tick)
}

func progressBody(extra string) string {
	return `{"assignment_id":"asn-a","task_id":"task-a",` + extra + `"daemon_id":"daemon-a","lines":[{"seq":1,"message":"thinking"},{"seq":2,"message":"done"}]}`
}
