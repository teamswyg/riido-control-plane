package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

const taskAgentGenerationSK = "GENERATION"

func dynamoDBTaskAgentPK(taskID, agentID string) string {
	identity, err := json.Marshal([2]string{taskID, agentID})
	if err != nil {
		panic(fmt.Sprintf("encode task Agent string identity: %v", err))
	}
	return "TASK_AGENT_EXECUTION#" + sha256Hex(identity)
}

func taskAgentDynamoKey(taskID, agentID, sk string) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": dynamoDBTaskAgentPK(taskID, agentID)}, "sk": {"S": sk}}
}

func taskAgentExecutionReceiptSK(assignment Assignment) string {
	return "EXEC#" + taskAgentAssignmentGeneration(assignment)
}

// Creation publishes its durable journal, runnable projection and membership
// observation together. Status updates never advance the generation marker.
func (s *DynamoDBAssignmentOperationStore) saveTaskAgentCreation(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	operation, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return err
	}
	projection, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return err
	}
	marker := taskAgentDynamoKey(record.TaskID, record.AgentID, taskAgentGenerationSK)
	marker["generation"] = map[string]string{"S": taskAgentAssignmentGeneration(record.Assignment)}
	receipt := taskAgentDynamoKey(record.TaskID, record.AgentID, taskAgentExecutionReceiptSK(record.Assignment))
	receipt["assignment_id"] = map[string]string{"S": record.AssignmentID}
	receipt["task_id"] = map[string]string{"S": record.TaskID}
	receipt["agent_id"] = map[string]string{"S": record.AgentID}
	items := []dynamoDBTransactWriteItem{}
	for i, item := range []map[string]map[string]string{operation, projection, marker, receipt} {
		condition := ""
		if i < 2 {
			condition = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
		}
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item, ConditionExpression: condition}})
	}
	payload, err := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{items})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err == nil {
		return nil
	}
	// Any failed response may follow a committed transaction. Confirm the exact
	// creation before reporting failure; a missing or different record is not success.
	journal, found, readErr := s.getTaskAgentItem(ctx, map[string]map[string]string{"pk": {"S": dynamoDBAssignmentOperationPK}, "sk": {"S": assignmentOperationSortKey(record)}}, credentials)
	if readErr != nil {
		return readErr
	}
	if found {
		previous, decodeErr := assignmentOperationRecordFromDynamoDBItem(journal)
		projection, exists, projectionErr := s.loadAssignmentProjection(ctx, record.AssignmentID, credentials)
		if decodeErr == nil && projectionErr == nil && exists && sameTaskAgentCreation(previous.Assignment, record.Assignment) && sameTaskAgentCreation(projection.Assignment, record.Assignment) {
			return nil
		}
	}
	return fmt.Errorf("dynamodb create task Agent execution: %w", err)
}
