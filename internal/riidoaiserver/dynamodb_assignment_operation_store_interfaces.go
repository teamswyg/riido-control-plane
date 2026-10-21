package riidoaiserver

var (
	_ AssignmentOperationStore   = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentOperationLoader  = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentQueueReader      = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentClaimer          = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentActiveLeaseStore = (*DynamoDBAssignmentOperationStore)(nil)
	_ AssignmentProjectionReader = (*DynamoDBAssignmentOperationStore)(nil)
)
