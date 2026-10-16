package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

// deviceConnectionGrantsReader reads only the saved device connection grants.
// It must not change snapshot bookkeeping, so a presence poll can run outside
// operationMu without affecting the hash fences of concurrent saves.
type deviceConnectionGrantsReader interface {
	loadDeviceConnectionGrants(ctx context.Context) ([]DeviceConnectionGrant, error)
}

// loadDeviceConnectionGrants gets the DEVICE_CONNECTION_GRANTS part directly
// instead of loading every snapshot part through the command loop. Its hash is
// not checked against the manifest: a part observed mid-write is still a real
// grant list, either the previous or the next one.
func (s *DynamoDBAIAgentClientSnapshot) loadDeviceConnectionGrants(ctx context.Context) ([]DeviceConnectionGrant, error) {
	credentials, err := cachedAWSCredentials(ctx, s.now, s.credentialsProvider, nil)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{s.tableName, true, map[string]map[string]string{
		"pk": {"S": dynamoDBAIAgentClientSnapshotPK},
		"sk": {"S": dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartDeviceConnections)},
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
		return nil, fmt.Errorf("decode DynamoDB device connection grants response: %w", err)
	}
	if len(response.Item) == 0 {
		return nil, nil
	}
	raw, err := gunzipBase64(dynamoDBStringValue(response.Item, "part_gzip"))
	if err != nil {
		return nil, fmt.Errorf("decode DynamoDB device connection grants part: %w", err)
	}
	var grants []DeviceConnectionGrant
	if err := json.Unmarshal(raw, &grants); err != nil {
		return nil, fmt.Errorf("decode DynamoDB device connection grants part: %w", err)
	}
	return grants, nil
}
