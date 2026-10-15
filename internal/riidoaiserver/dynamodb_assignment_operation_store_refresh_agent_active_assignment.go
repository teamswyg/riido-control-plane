package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) refreshAgentActiveAssignment(ctx context.Context, assignment Assignment, at time.Time, credentials AWSCredentials) error {
	if !assignmentHoldsActiveLease(assignment.State) {
		return nil
	}
	if strings.TrimSpace(assignment.AgentID) == "" || strings.TrimSpace(assignment.ID) == "" {
		return errors.New("riidoaiserver: active assignment refresh requires agent_id and assignment_id")
	}
	if strings.TrimSpace(assignment.LeaseToken) == "" {
		return errors.New("riidoaiserver: active assignment refresh requires lease_token")
	}
	payload, err := s.updateAgentActiveAssignmentPayload(assignment, at)
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     s.endpoint,
		endpointHost: s.endpointHost,
		region:       s.region,
		target:       dynamoDBUpdateItemTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   s.httpClient,
		now:          s.now,
	})
	if err == nil {
		return nil
	}
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("dynamodb update agent active assignment: %w", err)
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	if strings.Contains(dynamoErr.Type, "ConditionalCheckFailedException") {
		return nil
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = dynamoErr.MessageAlt
	}
	if dynamoErr.Message == "" {
		dynamoErr.Message = strings.TrimSpace(string(apiErr.body))
	}
	return fmt.Errorf("dynamodb update agent active assignment: status=%d type=%q message=%q", apiErr.statusCode, dynamoErr.Type, dynamoErr.Message)
}
