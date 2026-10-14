package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

const dynamoDBSnapshotTransactGetTarget = "DynamoDB_20120810.TransactGetItems"

// Query can interleave items from concurrent commits. Read the membership and
// its commit marker atomically before applying the manifest hash fence.
func (s *DynamoDBAIAgentClientSnapshot) readMembershipCommit(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) ([]map[string]map[string]string, error) {
	type get struct {
		TableName string                       `json:"TableName"`
		Key       map[string]map[string]string `json:"Key"`
	}
	type action struct {
		Get get `json:"Get"`
	}
	actions := make([]action, 0, 2)
	for _, sk := range []string{dynamoDBAIAgentClientSnapshotSK, dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)} {
		actions = append(actions, action{get{s.tableName, map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": sk}}}})
	}
	payload, err := json.Marshal(struct {
		TransactItems []action `json:"TransactItems"`
	}{actions})
	if err != nil {
		return nil, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region,
		target: dynamoDBSnapshotTransactGetTarget, payload: payload, credentials: credentials,
		httpClient: s.httpClient, now: s.now, traceAttrs: s.snapshotLoadTraceAttrs(),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb read AI Agent membership commit: %w", err)
	}
	var response struct {
		Responses []struct{ Item map[string]map[string]string } `json:"Responses"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if len(response.Responses) != 2 || len(response.Responses[0].Item) == 0 {
		return nil, fmt.Errorf("dynamodb membership commit response is incomplete")
	}
	result := make([]map[string]map[string]string, 0, len(items))
	for _, item := range items {
		if dynamoDBStringValue(item, "sk") != dynamoDBAIAgentClientSnapshotSK && dynamoDBAIAgentClientSnapshotPartName(item) != dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships {
			result = append(result, item)
		}
	}
	for _, read := range response.Responses {
		if len(read.Item) != 0 {
			result = append(result, read.Item)
		}
	}
	return result, nil
}
