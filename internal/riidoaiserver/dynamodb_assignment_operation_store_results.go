package riidoaiserver

type dynamoDBAssignmentOperationLoadResult struct {
	records []AssignmentOperationRecord
	err     error
}

type dynamoDBAssignmentQueueResult struct {
	assignments []Assignment
	err         error
}

type dynamoDBAssignmentClaimResult struct {
	result AssignmentClaimResult
	err    error
}

type dynamoDBAssignmentActiveLeaseResult struct {
	lease AssignmentActiveLease
	found bool
	err   error
}

type dynamoDBAssignmentProjectionResult struct {
	projection AssignmentProjection
	found      bool
	err        error
}
