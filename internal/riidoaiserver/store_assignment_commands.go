package riidoaiserver

import "context"

type assignCmd struct {
	creationReceipt           *taskAgentCreationReceipt
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
	durableProjection *AssignmentProjection
	AgentID           string
	AssignmentID      string
	Reason            string
}

type cancelAssignmentCmd struct {
	ctx    context.Context
	taskID string
	req    CancelAssignmentRequest
	reply  chan cancelAssignmentResult
}

type cancelAssignmentResult struct {
	assignment Assignment
	err        error
}

type pollCmd struct {
	ctx     context.Context
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
