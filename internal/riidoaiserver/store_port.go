package riidoaiserver

import "context"

type AssignmentStore interface {
	AssignTask(ctx context.Context, taskID string, req AssignRequest) (Assignment, error)
	AssignTaskAdditive(ctx context.Context, taskID string, req AssignRequest) (Assignment, error)
	PollAgent(ctx context.Context, agentID string, req PollRequest) (PollResponse, error)
	HeartbeatAgent(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, error)
	RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (AgentEventResponse, error)
	SubscribeTask(ctx context.Context, taskID string) ([]TaskEvent, <-chan TaskEvent, func(), error)
	MetricsReader
	Close()
}

type AssignmentCancellationStore interface {
	CancelAssignment(ctx context.Context, taskID string, req CancelAssignmentRequest) (Assignment, error)
}

type AssignmentHeartbeatEventStore interface {
	HeartbeatAgentWithEvents(ctx context.Context, agentID string, req AgentHeartbeatRequest) (AgentHeartbeatResponse, []TaskEvent, error)
}

type MetricsReader interface {
	Metrics(ctx context.Context) (MetricsSnapshot, error)
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
