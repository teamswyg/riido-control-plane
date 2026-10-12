package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"strconv"
)

var errAIAgentSnapshotConflict = errors.New("agent state changed concurrently; retry request")

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotTransaction(ctx context.Context, items []map[string]map[string]string, expected int64, credentials AWSCredentials) error {
	writes := make([]dynamoDBTransactWritePut, 0, len(items))
	for _, item := range items {
		var write dynamoDBTransactWritePut
		write.Put.TableName, write.Put.Item = s.tableName, item
		if dynamoDBStringValue(item, "sk") == dynamoDBAIAgentClientSnapshotSK {
			write.Put.ConditionExpression = "version = :expected"
			if expected == 0 {
				write.Put.ConditionExpression = "attribute_not_exists(version) OR version = :expected"
			}
			write.Put.ExpressionAttributeValues = map[string]map[string]string{":expected": {"N": strconv.FormatInt(expected, 10)}}
		}
		writes = append(writes, write)
	}
	payload, err := json.Marshal(struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}{writes})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now, traceAttrs: []TraceAttribute{StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), AIAgentClientSnapshotSave.String())}})
	if err != nil && isDynamoDBTransactionContention(err) {
		s.partHashes = nil
		return errAIAgentSnapshotConflict
	}
	return err
}
