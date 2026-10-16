package riidoaiserver

import (
	"context"
	"encoding/json"
)

// A presence read observes committed grants without changing snapshot fences.
// Atomic snapshot writes make a single part a committed old/new value.
type deviceConnectionGrantsReader interface {
	loadDeviceConnectionGrants(ctx context.Context) ([]DeviceConnectionGrant, error)
}

func (s *DynamoDBAIAgentClientSnapshot) loadDeviceConnectionGrants(ctx context.Context) ([]DeviceConnectionGrant, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case <-s.done:
		return nil, errDynamoDBAIAgentClientSnapshotClosed()
	default:
	}
	credentials, err := s.desktopPresenceCredentials(ctx)
	if err != nil {
		return nil, err
	}
	item, err := s.readDesktopPresenceItem(ctx, dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartDeviceConnections), credentials)
	if err != nil {
		return nil, err
	}
	if len(item) == 0 {
		return s.readLegacyDesktopPresence(ctx, credentials)
	}
	return decodeDesktopPresencePart(item)
}

func (s *DynamoDBAIAgentClientSnapshot) readDesktopPresenceItem(ctx context.Context, sk string, credentials AWSCredentials) (map[string]map[string]string, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{s.tableName, true, map[string]map[string]string{
		"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": sk},
	}})
	if err != nil {
		return nil, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region,
		target: dynamoDBGetItemTarget, payload: payload, credentials: credentials,
		httpClient: s.httpClient, now: s.now,
	})
	if err != nil {
		return nil, err
	}
	var response struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Item, nil
}
