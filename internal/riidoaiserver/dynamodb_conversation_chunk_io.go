package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAIAgentClientSnapshot) stageConversationChunks(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) ([]map[string]map[string]string, error) {
	writes := make([]map[string]map[string]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		if dynamoDBStringValue(item, "pk") != conversationChunkPK {
			writes = append(writes, item)
			continue
		}
		hash := dynamoDBStringValue(item, "sk")
		if !seen[hash] {
			if err := s.putSnapshotItem(ctx, item, credentials); err != nil {
				return nil, err
			}
			seen[hash] = true
		}
	}
	return writes, nil
}

func (s *DynamoDBAIAgentClientSnapshot) loadConversationChunks(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) ([]map[string]map[string]string, error) {
	seen := map[string]bool{}
	for _, item := range items {
		if dynamoDBStringValue(item, "pk") == conversationChunkPK {
			seen[dynamoDBStringValue(item, "sk")] = true
		}
	}
	missing := []string{}
	for _, item := range items {
		hashes, err := conversationChunkHashes(item)
		if err != nil {
			return nil, err
		}
		for _, hash := range hashes {
			if !seen[hash] {
				missing = append(missing, hash)
				seen[hash] = true
			}
		}
	}
	// Ten chunks stay below TransactGet's 4 MiB response budget.
	for offset := 0; offset < len(missing); offset += 10 {
		batch := missing[offset:min(offset+10, len(missing))]
		actions := make([]any, 0, len(batch))
		for _, hash := range batch {
			actions = append(actions, map[string]any{"Get": map[string]any{"TableName": s.tableName, "Key": map[string]map[string]string{"pk": {"S": conversationChunkPK}, "sk": {"S": hash}}}})
		}
		payload, err := json.Marshal(map[string]any{"TransactItems": actions})
		if err != nil {
			return nil, err
		}
		body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBSnapshotTransactGetTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err != nil {
			return nil, err
		}
		var result struct {
			Responses []struct{ Item map[string]map[string]string }
		}
		if json.Unmarshal(body, &result) != nil || len(result.Responses) != len(batch) {
			return nil, errors.New("incomplete conversation chunk response")
		}
		for i, response := range result.Responses {
			if dynamoDBStringValue(response.Item, "pk") != conversationChunkPK || dynamoDBStringValue(response.Item, "sk") != batch[i] {
				return nil, errors.New("conversation chunk identity mismatch")
			}
			items = append(items, response.Item)
		}
	}
	return items, nil
}
