package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func deviceBootstrapKey(id string) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": "DEVICE_BOOTSTRAP#" + id}, "sk": {"S": "TRANSITION"}}
}

func (s *DynamoDBAIAgentClientSnapshot) bootstrapGet(ctx context.Context, key map[string]map[string]string, credentials AWSCredentials) (map[string]map[string]string, error) {
	payload, err := json.Marshal(struct {
		TableName      string
		ConsistentRead bool
		Key            map[string]map[string]string
	}{s.tableName, true, key})
	if err != nil {
		return nil, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBGetItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return nil, err
	}
	var result struct{ Item map[string]map[string]string }
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Item, nil
}

func (s *DynamoDBAIAgentClientSnapshot) readDeviceBootstrap(ctx context.Context, id string) (deviceBootstrapRecord, bool, error) {
	var record deviceBootstrapRecord
	if !installationRandom32(id) {
		return record, false, errCLIInstallationInvalid
	}
	credentials, err := s.desktopPresenceCredentials(ctx)
	if err != nil {
		return record, false, err
	}
	item, err := s.bootstrapGet(ctx, deviceBootstrapKey(id), credentials)
	if err != nil || len(item) == 0 {
		return record, false, err
	}
	raw := dynamoDBStringValue(item, "payload_json")
	digest := sha256.Sum256([]byte(raw))
	if dynamoDBStringValue(item, "pk") != "DEVICE_BOOTSTRAP#"+id || dynamoDBStringValue(item, "sk") != "TRANSITION" || dynamoDBStringValue(item, "payload_sha256") != hex.EncodeToString(digest[:]) || json.Unmarshal([]byte(raw), &record) != nil || record.Request.TransitionID != id || record.valid() != nil {
		return deviceBootstrapRecord{}, false, errCLIInstallationUnavailable
	}
	return record, true, nil
}

func deviceBootstrapRecordAction(table string, mutation deviceBootstrapMutation) (dynamoDBTransactWriteItem, error) {
	item := deviceBootstrapKey(mutation.after.Request.TransitionID)
	raw, err := json.Marshal(mutation.after)
	if err != nil {
		return dynamoDBTransactWriteItem{}, err
	}
	digest := sha256.Sum256(raw)
	item["payload_json"], item["payload_sha256"] = map[string]string{"S": string(raw)}, map[string]string{"S": hex.EncodeToString(digest[:])}
	put := &dynamoDBTransactWritePutAction{TableName: table, Item: item, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
	if mutation.before != nil {
		old, err := json.Marshal(mutation.before)
		if err != nil {
			return dynamoDBTransactWriteItem{}, err
		}
		hash := sha256.Sum256(old)
		put.ConditionExpression = "payload_sha256 = :bootstrap_hash"
		put.ExpressionAttributeValues = map[string]map[string]string{":bootstrap_hash": {"S": hex.EncodeToString(hash[:])}}
	}
	return dynamoDBTransactWriteItem{Put: put}, nil
}
