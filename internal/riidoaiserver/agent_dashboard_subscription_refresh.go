package riidoaiserver

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

// Local transitions are pushed immediately. On keepalive, all subscribers of
// an idle replica share one revision check; an unchanged revision needs no replay.
func (s *PersistentAIAgentClientStore) refreshDashboardSubscription(ctx context.Context) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	now := s.snapshotNow()
	if !s.dashboardLastRefresh.IsZero() && now.Sub(s.dashboardLastRefresh) < sseKeepaliveInterval {
		return nil
	}
	if reader, ok := s.snapshotStore.(interface {
		dashboardSnapshotVersion(context.Context) (int64, error)
	}); ok {
		version, err := reader.dashboardSnapshotVersion(ctx)
		if err != nil {
			return err
		}
		s.snapshotReloadMu.Lock()
		required := s.snapshotReloadRequired
		s.snapshotReloadMu.Unlock()
		if version == s.version && !required {
			s.dashboardLastRefresh = now
			return nil
		}
	}
	s.snapshotReloadMu.Lock()
	s.snapshotLastReloadAt = time.Time{}
	s.snapshotReloadMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	s.dashboardLastRefresh = now
	return nil
}

func (s *DynamoDBAIAgentClientSnapshot) dashboardSnapshotVersion(ctx context.Context) (int64, error) {
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return 0, err
	}
	payload, err := json.Marshal(map[string]any{"TableName": s.tableName, "ConsistentRead": true,
		"Key":                  map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": dynamoDBAIAgentClientSnapshotSK}},
		"ProjectionExpression": "#v", "ExpressionAttributeNames": map[string]string{"#v": "version"}})
	if err != nil {
		return 0, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBGetItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return 0, err
	}
	var response struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}
	if len(response.Item) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(dynamoDBNumberValue(response.Item, "version"), 10, 64)
}
