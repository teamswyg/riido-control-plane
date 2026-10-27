package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// Strong base-table reads; bounded pagination never turns incomplete history into safe idle.
func (s *DynamoDBAssignmentOperationStore) loadDeviceExecutionClaims(ctx context.Context, deviceID string) ([]executionClaim, error) {
	if strings.TrimSpace(deviceID) == "" || len(deviceID) > 512 {
		return nil, errors.New("execution claim device is required")
	}
	pk := "DEVICE_EXECUTION#" + deviceID
	var claims []executionClaim
	var start map[string]map[string]string
	seen := map[string]bool{}
	for range 100 {
		body, err := s.assignmentContextRequest(ctx, dynamoDBQueryTarget, struct {
			TableName                 string
			ConsistentRead            bool
			KeyConditionExpression    string
			ExpressionAttributeValues map[string]map[string]string
			ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
			Limit                     int
		}{s.tableName, true, "pk = :pk", map[string]map[string]string{":pk": {"S": pk}}, start, 100})
		if err != nil {
			return nil, err
		}
		var page struct {
			Items            []map[string]map[string]string
			LastEvaluatedKey map[string]map[string]string
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		if len(page.Items) > 100 {
			return nil, errors.New("execution claim page exceeds bound")
		}
		for _, item := range page.Items {
			claim, err := executionClaimFromItem(item)
			if err != nil || claim.DeviceID != deviceID || seen[claim.key()] {
				return nil, errors.New("invalid execution claim page")
			}
			seen[claim.key()] = true
			claims = append(claims, claim)
		}
		if len(page.LastEvaluatedKey) == 0 {
			return claims, nil
		}
		next := dynamoDBStringValue(page.LastEvaluatedKey, "sk")
		if dynamoDBStringValue(page.LastEvaluatedKey, "pk") != pk || next == "" || next <= dynamoDBStringValue(start, "sk") {
			return nil, errors.New("execution claim pagination did not progress")
		}
		start = page.LastEvaluatedKey
	}
	return nil, errors.New("execution claim history exceeds bounded read")
}
