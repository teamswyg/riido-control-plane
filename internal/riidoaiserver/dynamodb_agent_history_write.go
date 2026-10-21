package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func agentHistoryPK(agentID string) string { return "AGENT_HISTORY#" + sha256Hex([]byte(agentID)) }
func (s *DynamoDBAssignmentOperationStore) saveAgentHistoryOperation(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	item, err := agentHistoryOperationItem(record)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		TableName           string
		Item                map[string]map[string]string
		ConditionExpression string
	}{s.tableName, item, "attribute_not_exists(pk) AND attribute_not_exists(sk)"})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBPutItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	var apiErr dynamoDBAPIError
	if errors.As(err, &apiErr) && strings.Contains(string(apiErr.body), "ConditionalCheckFailedException") {
		return nil
	}
	return err
}

func agentHistoryOperationItem(record AssignmentOperationRecord) (map[string]map[string]string, error) {
	item, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return nil, err
	}
	item["pk"] = map[string]string{"S": agentHistoryPK(record.AgentID)}
	item["sk"] = map[string]string{"S": record.OperationID}
	return item, nil
}
