package riidoaiserver

import (
	"context"
	"time"
)

type dynamoDBAssignmentOperationCommand struct {
	ctx             context.Context
	load            bool
	queue           bool
	claim           bool
	active          bool
	refresh         bool
	projection      bool
	agentID         string
	assignmentID    string
	claimAt         time.Time
	refreshAt       time.Time
	record          *AssignmentOperationRecord
	assignment      *Assignment
	close           bool
	reply           chan error
	loadReply       chan dynamoDBAssignmentOperationLoadResult
	queueReply      chan dynamoDBAssignmentQueueResult
	claimReply      chan dynamoDBAssignmentClaimResult
	activeReply     chan dynamoDBAssignmentActiveLeaseResult
	projectionReply chan dynamoDBAssignmentProjectionResult
}
