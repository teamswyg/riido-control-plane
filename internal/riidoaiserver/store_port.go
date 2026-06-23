package riidoaiserver

import (
	"context"
	"time"
)

type AssignmentStore interface {
	AssignTask(ctx context.Context, taskID string, req AssignRequest) (Assignment, error)
	AssignTaskReplacement(ctx context.Context, taskID string, req AssignRequest) (Assignment, error)
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

// AssignmentLongPollStore is the optional long-poll claim capability. A store
// that implements it lets the poll handler hold a request (PollRequest.WaitMs)
// until work is available or the budget elapses. A store that does not implement
// it transparently degrades to the point-in-time PollAgent path.
type AssignmentLongPollStore interface {
	WaitForAssignment(ctx context.Context, agentID string, req PollRequest, hold, tick time.Duration) (PollResponse, error)
}

type AssignmentToolApprovalStore interface {
	CreateToolApproval(ctx context.Context, agentID string, req ToolApprovalRequest) (ToolApprovalRequest, error)
	DecideToolApproval(ctx context.Context, taskID string, decision ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error)
	ListTaskToolApprovals(ctx context.Context, taskID string) ([]ToolApprovalRequest, error)
	WaitForToolApproval(ctx context.Context, agentID, assignmentID, approvalID string, hold, tick time.Duration) (ToolApprovalResult, *ToolApprovalDecision, error)
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
