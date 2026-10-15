package riidoaiserver

import (
	"context"
	"encoding/json"
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) saveConditionalAssignment(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	operation, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return err
	}
	projection, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return err
	}
	expected := record.ExpectedProjection
	items := make([]dynamoDBTransactWritePut, 2)
	for i, item := range []map[string]map[string]string{projection, operation} {
		items[i].Put.TableName = s.tableName
		items[i].Put.Item = item
	}
	items[0].Put.ConditionExpression = "assignment_state = :expected_state AND last_event_seq = :expected_seq"
	items[0].Put.ExpressionAttributeValues = map[string]map[string]string{
		":expected_state": {"S": string(expected.Assignment.State)},
		":expected_seq":   {"N": strconv.FormatInt(expected.LastEventSeq, 10)},
	}
	if expected.Assignment.LeaseToken != "" {
		items[0].Put.ConditionExpression += " AND lease_token = :expected_lease"
		items[0].Put.ExpressionAttributeValues[":expected_lease"] = map[string]string{"S": expected.Assignment.LeaseToken}
	}
	items[1].Put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
	payload, err := json.Marshal(struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{items})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil && isDynamoDBTransactionContention(err) {
		return errAssignmentProjectionConflict
	}
	if err != nil {
		return err
	}
	// Keep the lease while cancelling so another request cannot start before ACK.
	if err := s.releaseAgentActiveAssignmentIfTerminal(ctx, record, credentials); err != nil {
		return err
	}
	return s.deleteTaskAgentTerminalReceipt(ctx, record, credentials)
}
