package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAssignmentOperationStore) saveIdentifiedAssignment(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	operation, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		return err
	}
	operation["sk"] = map[string]string{"S": "request#" + record.AssignmentID}
	projection, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		return err
	}
	items := make([]dynamoDBTransactWritePut, 2)
	for i, item := range []map[string]map[string]string{projection, operation} {
		items[i].Put.TableName = s.tableName
		items[i].Put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
		items[i].Put.Item = item
	}
	payload, err := json.Marshal(struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{items})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil && isDynamoDBTransactionContention(err) {
		_, found, readErr := s.loadAssignmentProjection(ctx, record.AssignmentID, credentials)
		if readErr != nil {
			return readErr
		}
		if found {
			return nil
		}
	}
	return err
}
