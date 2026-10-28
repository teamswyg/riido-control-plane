package riidoaiserver

import "time"

// ExecutionCoverageAuthorityCompatible checks the native atomic table boundary.
// Only the shared factory establishes the private authority identity.
func ExecutionCoverageAuthorityCompatible(snapshot AIAgentClientSnapshotStore, assignments AssignmentOperationStore) bool {
	clients, ok := snapshot.(*DynamoDBAIAgentClientSnapshot)
	if !ok || clients == nil {
		return false
	}
	operations, ok := assignments.(*DynamoDBAssignmentOperationStore)
	return ok && operations != nil && operations.executionCoverageAuthority == clients && clients.tableName == operations.tableName && clients.region == operations.region && clients.endpoint == operations.endpoint
}

// NewExecutionCoverageAssignmentStore shares the exact credential provider and
// native table authority at construction, before either adapter accepts work.
func NewExecutionCoverageAssignmentStore(snapshot AIAgentClientSnapshotStore, activeLease time.Duration) (*DynamoDBAssignmentOperationStore, error) {
	clients, ok := snapshot.(*DynamoDBAIAgentClientSnapshot)
	if !ok || clients == nil {
		return nil, errCLIInstallationUnavailable
	}
	operations, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{
		EnableAssignmentContextRetention: true, EnableAgentHistory: true,
		Region: clients.region, TableName: clients.tableName, Endpoint: clients.endpoint,
		HTTPClient: clients.httpClient, Now: clients.now, ActiveLeaseDuration: activeLease,
		CredentialsProvider: clients.credentialsProvider,
	})
	if err != nil {
		return nil, err
	}
	operations.executionCoverageAuthority = clients
	return operations, nil
}
