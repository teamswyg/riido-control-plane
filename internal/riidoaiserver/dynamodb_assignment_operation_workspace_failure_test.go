package riidoaiserver

import "testing"

func TestDynamoDBAssignmentOperationStoreLoadsWorkspaceFailureFromTerminalOperation(t *testing.T) {
	testDynamoDBTerminalAssignmentEvent(t, AssignmentFailed, EventAssignmentFailed)
}

func TestDynamoDBAssignmentOperationStoreLoadsCompletedResultFromTerminalOperation(t *testing.T) {
	testDynamoDBTerminalAssignmentEvent(t, AssignmentCompleted, EventAssignmentCompleted)
}
