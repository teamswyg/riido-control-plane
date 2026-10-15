package riidoaiserver

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) deleteTaskAgentTerminalReceipt(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	// Legacy assignment identities predate the receipt index.
	if !isTerminal(record.Assignment.State) || len(record.AssignmentID) != 36 || !strings.HasPrefix(record.AssignmentID, "asn-") {
		return nil
	}
	projection, found, err := s.loadAssignmentProjection(ctx, record.AssignmentID, credentials)
	if err != nil {
		return err
	}
	if !found || !isTerminal(projection.Assignment.State) {
		return nil
	}
	actions := []dynamoDBTransactWriteItem{
		{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": dynamoDBAssignmentProjectionPK(record.AssignmentID)}, "sk": {"S": dynamoDBAssignmentProjectionSK}}, ConditionExpression: "assignment_state = :terminal AND last_event_seq = :terminal_seq", ExpressionAttributeValues: map[string]map[string]string{":terminal": {"S": string(projection.Assignment.State)}, ":terminal_seq": {"N": strconv.FormatInt(projection.LastEventSeq, 10)}}}},
		{Delete: &dynamoDBTransactWriteDeleteAction{TableName: s.tableName, Key: taskAgentDynamoKey(record.TaskID, record.AgentID, taskAgentExecutionReceiptSK(record.Assignment))}},
	}
	payload, err := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if isDynamoDBTransactionContention(err) {
		return nil
	}
	return err
}
