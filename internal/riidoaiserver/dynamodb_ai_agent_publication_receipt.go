package riidoaiserver

import "context"

// The receipt exists only while execution is active. Couple its publication proof
// to the membership commit, so later failed requests cannot cancel accepted runs.
func (s *DynamoDBAIAgentClientSnapshot) publicationReceiptAction(ctx context.Context, mutation taskAgentMembershipSnapshotMutation, credentials AWSCredentials) (*dynamoDBTransactWriteItem, error) {
	if mutation.publicationAssignmentID == "" {
		return nil, nil
	}
	reader := &DynamoDBAssignmentOperationStore{tableName: s.tableName, region: s.region, endpoint: s.endpoint, endpointHost: s.endpointHost, httpClient: s.httpClient, now: s.now}
	scope := mutation.scope
	item, found, err := reader.getTaskAgentItem(ctx, taskAgentDynamoKey(scope.taskID, scope.agentID, "EXEC#"+scope.generation), credentials)
	if err != nil {
		return nil, err
	}
	if !found {
		projection, exists, err := reader.loadAssignmentProjection(ctx, mutation.publicationAssignmentID, credentials)
		if err != nil {
			return nil, err
		}
		if !exists || projection.Assignment.TaskID != scope.taskID || projection.Assignment.AgentID != scope.agentID || taskAgentAssignmentGeneration(projection.Assignment) != scope.generation || !isTerminal(projection.Assignment.State) {
			return nil, ErrAIAgentMembershipChanged
		}
		// Terminal states cannot become runnable again under the same identity.
		return nil, nil
	}
	item["publication_committed"] = map[string]string{"S": "true"}
	return &dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item, ConditionExpression: "attribute_exists(pk) AND attribute_exists(sk)"}}, nil
}
