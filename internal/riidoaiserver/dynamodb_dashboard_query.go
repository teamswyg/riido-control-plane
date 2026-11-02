package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) dashboardItems(ctx context.Context, credentials AWSCredentials, agentID, prefix, fields string, visit func(map[string]map[string]string) error) error {
	query := struct {
		TableName, KeyConditionExpression, ProjectionExpression string
		ExpressionAttributeValues                               map[string]map[string]string
		ConsistentRead                                          bool
		ExclusiveStartKey                                       map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
	}{s.tableName, "pk = :pk AND begins_with(sk, :prefix)", fields, map[string]map[string]string{":pk": {"S": agentHistoryPK(agentID)}, ":prefix": {"S": prefix}}, true, nil}
	for {
		payload, err := json.Marshal(query)
		if err != nil {
			return err
		}
		body, err := loadDynamoDBAssignmentOperationPage(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBQueryTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err != nil {
			return err
		}
		var page struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err = json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, item := range page.Items {
			if err := visit(item); err != nil {
				return err
			}
		}
		if len(page.LastEvaluatedKey) == 0 {
			return nil
		}
		query.ExclusiveStartKey = page.LastEvaluatedKey
		if err := paceDynamoDBAssignmentOperationReplay(ctx); err != nil {
			return err
		}
	}
}

func (s *DynamoDBAssignmentOperationStore) dashboardMemberIDs(ctx context.Context, credentials AWSCredentials, agentID string, ids map[string]bool) error {
	return s.dashboardItems(ctx, credentials, agentID, "assign:", "assignment_id", func(item map[string]map[string]string) error {
		id := dynamoDBStringValue(item, "assignment_id")
		if id == "" {
			return fmt.Errorf("missing dashboard assignment id")
		}
		ids[id] = true
		return nil
	})
}
