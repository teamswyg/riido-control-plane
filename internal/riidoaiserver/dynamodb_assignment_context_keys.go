package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
)

// Assignment context items live under their own partition and never carry
// agent_id/assignment_sort, so they stay out of the agent_queue GSI.
const (
	dynamoDBAssignmentContextPKPrefix   = "ASSIGNMENT_CONTEXT#"
	dynamoDBAssignmentContextSnapshotSK = "SNAPSHOT"
	dynamoDBAssignmentContextReadPrefix = "READ#"
)

func dynamoDBAssignmentContextPK(assignmentID string) string {
	return dynamoDBAssignmentContextPKPrefix + strings.TrimSpace(assignmentID)
}

const (
	dynamoDBAssignmentContextTimeFormat = "2006-01-02T15:04:05.000000000Z07:00"
	dynamoDBAssignmentContextReadBatch  = 100
)

func (s *DynamoDBAssignmentOperationStore) assignmentContextRequest(ctx context.Context, target string, request any) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return nil, err
	}
	return doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: target, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
}
