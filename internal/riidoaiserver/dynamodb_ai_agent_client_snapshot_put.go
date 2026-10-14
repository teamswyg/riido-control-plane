package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotItem(ctx context.Context, item map[string]map[string]string, credentials AWSCredentials) error {
	payload, err := json.Marshal(struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}{
		TableName: s.tableName,
		Item:      item,
	})
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
		traceAttrs: []TraceAttribute{
			StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), AIAgentClientSnapshotSave.String()),
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb save AI Agent client snapshot: %w", err)
	}
	return nil
}
