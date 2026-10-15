package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *DynamoDBAIAgentClientSnapshot) putMembershipCommit(ctx context.Context, membership, manifest map[string]map[string]string, credentials AWSCredentials) error {
	type put struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}
	type action struct {
		Put put `json:"Put"`
	}
	payload, err := json.Marshal(struct {
		TransactItems []action `json:"TransactItems"`
	}{[]action{{put{s.tableName, membership}}, {put{s.tableName, manifest}}}})
	if err != nil {
		return err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region,
		target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials,
		httpClient: s.httpClient, now: s.now,
		traceAttrs: []TraceAttribute{StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), AIAgentClientSnapshotSave.String())},
	})
	if err != nil {
		return fmt.Errorf("dynamodb commit AI Agent membership snapshot: %w", err)
	}
	return nil
}
