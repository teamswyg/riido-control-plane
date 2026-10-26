package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error {
	if strings.TrimSpace(snapshot.AssignmentID) == "" {
		return errors.New("assignment_id is required")
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	compressed, err := gzipBase64(encoded)
	if err != nil {
		return err
	}
	item := map[string]map[string]string{
		"pk":             {"S": dynamoDBAssignmentContextPK(snapshot.AssignmentID)},
		"sk":             {"S": dynamoDBAssignmentContextSnapshotSK},
		"schema":         {"S": "assignment_context_snapshot/v1"},
		"payload_gzip":   {"S": compressed},
		"context_task":   {"S": snapshot.TaskID},
		"context_agent":  {"S": snapshot.AgentID},
		"context_stored": {"S": snapshot.CapturedAt.Format(dynamoDBAssignmentContextTimeFormat)},
	}
	_, err = s.assignmentContextRequest(ctx, dynamoDBPutItemTarget, struct {
		TableName string
		Item      map[string]map[string]string
	}{s.tableName, item})
	return err
}

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentContextSnapshot(ctx context.Context, assignmentID string) (AssignmentContextSnapshot, bool, error) {
	body, err := s.assignmentContextRequest(ctx, dynamoDBGetItemTarget, struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, map[string]map[string]string{
		"pk": {"S": dynamoDBAssignmentContextPK(assignmentID)},
		"sk": {"S": dynamoDBAssignmentContextSnapshotSK},
	}})
	if err != nil {
		return AssignmentContextSnapshot{}, false, err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil || len(response.Item) == 0 {
		return AssignmentContextSnapshot{}, false, err
	}
	decoded, err := gunzipBase64(response.Item["payload_gzip"]["S"])
	if err != nil {
		return AssignmentContextSnapshot{}, false, fmt.Errorf("decode assignment context snapshot: %w", err)
	}
	var snapshot AssignmentContextSnapshot
	if err := json.Unmarshal(decoded, &snapshot); err != nil || snapshot.AssignmentID != strings.TrimSpace(assignmentID) {
		return AssignmentContextSnapshot{}, false, errors.New("assignment context snapshot scope mismatch")
	}
	return snapshot, true, nil
}
