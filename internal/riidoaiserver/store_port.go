package riidoaiserver

import "context"

type AssignmentStore interface {
	AssignTask(ctx context.Context, taskID string, req AssignRequest) (Assignment, error)
	PollAgent(ctx context.Context, agentID string, req PollRequest) (PollResponse, error)
	HeartbeatAgent(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error)
	RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (AgentEventResponse, error)
	SubscribeTask(ctx context.Context, taskID string) ([]TaskEvent, <-chan TaskEvent, func(), error)
	Metrics(ctx context.Context) (MetricsSnapshot, error)
	Close()
}

type EventSink interface {
	AppendTaskEvent(ctx context.Context, event TaskEvent) error
	Close() error
}

type SnapshotStore interface {
	LoadStoreSnapshot(ctx context.Context) (StoreSnapshot, bool, error)
	SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error
	Close() error
}
