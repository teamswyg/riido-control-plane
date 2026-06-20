package riidoaiserver

type subscribeCmd struct {
	taskID string
	reply  chan subscribeResult
}

type subscribeResult struct {
	history []TaskEvent
	events  <-chan TaskEvent
	subID   int64
	err     error
}

type unsubscribeCmd struct {
	taskID string
	subID  int64
	reply  chan struct{}
}

type registerWaiterCmd struct {
	agentID string
	reply   chan registerWaiterResult
}

type registerWaiterResult struct {
	ch chan struct{}
	id int64
}

type unregisterWaiterCmd struct {
	agentID string
	id      int64
	reply   chan struct{}
}

type metricsCmd struct {
	reply chan metricsResult
}

type metricsResult struct {
	snapshot MetricsSnapshot
	err      error
}

type assignmentProjectionCmd struct {
	assignmentID string
	reply        chan assignmentProjectionResult
}

type assignmentProjectionResult struct {
	projection AssignmentProjection
	found      bool
	err        error
}

type closeCmd struct {
	reply chan struct{}
}
