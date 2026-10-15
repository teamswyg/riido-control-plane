package riidoaiserver

import (
	"context"
	"encoding/json"
)

func assignmentProjectionRecoveryField(item map[string]map[string]string, record AssignmentOperationRecord) (map[string]map[string]string, error) {
	watermark, err := recoveryWatermarkForRecord(record)
	if err != nil || watermark == nil {
		return item, err
	}
	encoded, err := json.Marshal(watermark)
	if err != nil {
		return nil, err
	}
	item["recovery_watermark_json"] = map[string]string{"S": string(encoded)}
	return item, nil
}

// Legacy stale-lease repair writes must preserve a journal committed elsewhere.
// Only a conflicting projection needs this read; ordinary lease heartbeats do not.
func (s *DynamoDBAssignmentOperationStore) preserveRecoveryProjection(ctx context.Context, record AssignmentOperationRecord, credentials AWSCredentials) error {
	for range 3 {
		current, found, err := s.loadAssignmentProjection(ctx, record.AssignmentID, credentials)
		if err != nil || !found || current.LastEventSeq > assignmentOperationLastEventSeq(record) || current.Recovery == nil {
			return err
		}
		expected := AssignmentProjection(current)
		record.ExpectedProjection = &expected
		payload, err := s.putAssignmentProjectionPayload(record)
		if err != nil {
			return err
		}
		_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBPutItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err == nil {
			return nil
		}
		if !isDynamoDBTransactionContention(err) {
			return err
		}
	}
	return errAssignmentProjectionConflict
}
