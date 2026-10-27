package riidoaiserver

import (
	"encoding/json"
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) claimTransactionPayload(record AssignmentOperationRecord, expectedLastEventSeq int64) ([]byte, error) {
	operationItem, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	projectionItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	payload := struct {
		TransactItems []dynamoDBTransactWriteItem `json:"TransactItems"`
	}{TransactItems: []dynamoDBTransactWriteItem{{Put: &dynamoDBTransactWritePutAction{}}, {Put: &dynamoDBTransactWritePutAction{}}, {Put: &dynamoDBTransactWritePutAction{}}}}
	payload.TransactItems[0].Put.TableName = s.tableName
	payload.TransactItems[0].Put.ConditionExpression = "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq"
	payload.TransactItems[0].Put.ExpressionAttributeValues = map[string]map[string]string{
		":queued":                  {"S": string(AssignmentQueued)},
		":agent_id":                {"S": record.AgentID},
		":expected_last_event_seq": {"N": strconv.FormatInt(expectedLastEventSeq, 10)},
	}
	payload.TransactItems[0].Put.Item = projectionItem
	payload.TransactItems[1].Put.TableName = s.tableName
	payload.TransactItems[1].Put.ConditionExpression = "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms"
	payload.TransactItems[1].Put.ExpressionAttributeValues = map[string]map[string]string{
		":claim_started_unix_ms": {"N": strconv.FormatInt(record.RecordedAt.UTC().UnixMilli(), 10)},
	}
	payload.TransactItems[1].Put.Item = s.agentActiveAssignmentDynamoDBItem(record)
	payload.TransactItems[2].Put.TableName = s.tableName
	payload.TransactItems[2].Put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
	payload.TransactItems[2].Put.Item = operationItem
	if record.ExecutionClaim != nil {
		item, err := executionClaimItem(*record.ExecutionClaim)
		if err != nil {
			return nil, err
		}
		sidecar := dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)", Item: item}}
		payload.TransactItems = append(payload.TransactItems, sidecar, s.devicePolicyAdvanceAction(record.ExecutionClaim.DeviceID))
	}
	return json.Marshal(payload)
}
