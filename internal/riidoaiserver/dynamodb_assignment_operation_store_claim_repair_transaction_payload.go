package riidoaiserver

import (
	"encoding/json"
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) claimRepairTransactionPayload(record AssignmentOperationRecord, expectedLastEventSeq int64, repairs []dynamoDBAssignmentClaimRepair, activeCondition *dynamoDBTransactWriteConditionCheckAction) ([]byte, error) {
	items := []dynamoDBTransactWriteItem{}
	if activeCondition != nil {
		items = append(items, dynamoDBTransactWriteItem{ConditionCheck: activeCondition})
	}
	for _, repair := range repairs {
		projectionItem, err := assignmentProjectionDynamoDBItem(repair.Operation)
		if err != nil {
			return nil, err
		}
		operationItem, err := assignmentOperationDynamoDBItem(repair.Operation)
		if err != nil {
			return nil, err
		}
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
			TableName:           s.tableName,
			ConditionExpression: "assignment_state = :expected_state AND last_event_seq = :expected_last_event_seq",
			ExpressionAttributeValues: map[string]map[string]string{
				":expected_state":          {"S": string(repair.ExpectedState)},
				":expected_last_event_seq": {"N": strconv.FormatInt(repair.ExpectedLastEventSeq, 10)},
			},
			Item: projectionItem,
		}})
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
			TableName:           s.tableName,
			ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
			Item:                operationItem,
		}})
	}
	projectionItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	operationItem, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq",
		ExpressionAttributeValues: map[string]map[string]string{
			":queued":                  {"S": string(AssignmentQueued)},
			":agent_id":                {"S": record.AgentID},
			":expected_last_event_seq": {"N": strconv.FormatInt(expectedLastEventSeq, 10)},
		},
		Item: projectionItem,
	}})
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms",
		ExpressionAttributeValues: map[string]map[string]string{
			":claim_started_unix_ms": {"N": strconv.FormatInt(record.RecordedAt.UTC().UnixMilli(), 10)},
		},
		Item: s.agentActiveAssignmentDynamoDBItem(record),
	}})
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName:           s.tableName,
		ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)",
		Item:                operationItem,
	}})
	items, err = s.appendExecutionClaimAction(items, record.ExecutionClaim, record.executionAdmission)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		TransactItems []dynamoDBTransactWriteItem `json:"TransactItems"`
	}{items})
}
