package riidoaiserver

import "time"

type Store struct {
	commands            chan any
	done                chan struct{}
	now                 func() time.Time
	activeLeaseDuration time.Duration
	outbox              EventSink
	snapshotStore       SnapshotStore
	operationStore      AssignmentOperationStore
	agentRegistry       AgentRegistry
	operationMetrics    *StoreOperationMetrics
	traceRecorder       TraceRecorder
}

type StoreConfig struct {
	Now                 func() time.Time
	ActiveLeaseDuration time.Duration
	Outbox              EventSink
	SnapshotStore       SnapshotStore
	OperationStore      AssignmentOperationStore
	AgentRegistry       AgentRegistry
	OperationMetrics    *StoreOperationMetrics
	TraceRecorder       TraceRecorder
}
