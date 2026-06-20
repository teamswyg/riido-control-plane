package riidoaiserver

import "time"

func NewStore() *Store {
	return NewStoreWithConfig(StoreConfig{})
}

func NewStoreWithClock(now func() time.Time) *Store {
	return NewStoreWithConfig(StoreConfig{Now: now})
}

func NewStoreWithConfig(config StoreConfig) *Store {
	return newStoreWithConfig(config, newStoreState())
}

func newStoreWithConfig(config StoreConfig, initial storeState) *Store {
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	activeLeaseDuration := config.ActiveLeaseDuration
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	s := &Store{
		commands:            make(chan any, 64),
		done:                make(chan struct{}),
		now:                 now,
		activeLeaseDuration: activeLeaseDuration,
		outbox:              config.Outbox,
		snapshotStore:       config.SnapshotStore,
		operationStore:      config.OperationStore,
		agentRegistry:       config.AgentRegistry,
		operationMetrics:    config.OperationMetrics,
		traceRecorder:       config.TraceRecorder,
	}
	if s.operationMetrics == nil {
		s.operationMetrics = NewStoreOperationMetrics()
	}
	go s.loop(initial)
	return s
}
