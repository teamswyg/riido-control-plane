package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) save(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) (err error) {
	defer func() {
		if err == nil && s.enableAgentHistory && record.ExpectedProjection == nil && !record.createsExecution && (record.OperationType != AssignmentOperationAssignTask || record.Assignment.RequestID == "") {
			err = s.saveAgentHistoryOperation(ctx, record, credentials)
		}
	}()
	if record.SchemaVersion == "" {
		record.SchemaVersion = AssignmentOperationSchemaVersion
	}
	if record.RecordedAt.IsZero() {
		record.RecordedAt = s.now()
	}
	if err := validateAssignmentOperationRecord(record); err != nil {
		return err
	}
	if record.ExpectedProjection != nil {
		return s.saveConditionalAssignment(ctx, record, credentials)
	}
	if record.createsExecution || (record.OperationType == AssignmentOperationAssignTask && record.Assignment.RequestID != "") {
		return s.saveTaskAgentCreation(ctx, record, credentials)
	}
	payload, err := s.putOperationPayload(record)
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
		return s.saveAssignmentProjection(ctx, record, credentials)
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb put assignment operation: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return s.saveAssignmentProjection(ctx, record, credentials)
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb put assignment operation: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}
