package riidoaiserver

type assignCmd struct {
	taskID                    string
	req                       AssignRequest
	allowConcurrentTaskAgents bool
	forceNewAssignment        bool
	reply                     chan assignResult
}

type assignResult struct {
	assignment Assignment
	err        error
}

type CancelAssignmentRequest struct {
	AgentID      string
	AssignmentID string
	Reason       string
}

type cancelAssignmentCmd struct {
	taskID string
	req    CancelAssignmentRequest
	reply  chan cancelAssignmentResult
}

type cancelAssignmentResult struct {
	assignment Assignment
	err        error
}

type pollCmd struct {
	agentID string
	req     PollRequest
	// countRequest increments pollRequestsTotal for this evaluation. A daemon
	// long-poll counts exactly once (its first evaluation) so the metric still
	// means "daemon poll requests", not internal re-evaluations during a hold.
	countRequest bool
	reply        chan pollResult
}

type pollResult struct {
	response              PollResponse
	operationAlreadySaved bool
	mutatedAssignment     *Assignment
	mutationOperationType AssignmentOperationType
	err                   error
}
