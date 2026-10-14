package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) saveAssignmentProjection(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	payload, err := s.putAssignmentProjectionPayload(record)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBPutItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		if err := s.releaseAgentActiveAssignmentIfTerminal(ctx, record, credentials); err != nil {
			return err
		}
		return s.deleteTaskAgentTerminalReceipt(ctx, record, credentials)
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb put assignment projection: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		if err := s.releaseAgentActiveAssignmentIfTerminal(ctx, record, credentials); err != nil {
			return err
		}
		return s.deleteTaskAgentTerminalReceipt(ctx, record, credentials)
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb put assignment projection: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}
