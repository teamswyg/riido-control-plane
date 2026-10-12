package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *DynamoDBAIAgentClientSnapshot) load(ctx context.Context, credentials AWSCredentials) (snapshot AIAgentClientSnapshot, ok bool, err error) {
	startedAt := time.Now()
	var snapshotBytes int64
	defer func() {
		s.observeAIAgentClientSnapshot(AIAgentClientSnapshotLoad, startedAt, snapshotBytes, err)
	}()
	return s.loadConsistentSnapshot(ctx, credentials, &snapshotBytes)
}

func (s *DynamoDBAIAgentClientSnapshot) loadConsistentSnapshot(ctx context.Context, credentials AWSCredentials, snapshotBytes *int64) (AIAgentClientSnapshot, bool, error) {
	for range 3 {
		items, err := s.querySnapshotItems(ctx, credentials)
		if err != nil {
			return AIAgentClientSnapshot{}, false, err
		}
		current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
		if len(current) == 0 {
			return AIAgentClientSnapshot{}, false, nil
		}
		if dynamoDBAIAgentClientSnapshotItemIsLegacy(current) {
			return s.loadLegacySnapshot(ctx, current, credentials)
		}
		snapshot, size, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
		*snapshotBytes = size
		if errors.Is(err, errAIAgentSnapshotConflict) {
			continue
		}
		if err != nil {
			return AIAgentClientSnapshot{}, false, err
		}
		s.partHashes = dynamoDBAIAgentClientSnapshotPartHashes(items)
		return snapshot, true, nil
	}
	return AIAgentClientSnapshot{}, false, errAIAgentSnapshotConflict
}

func (s *DynamoDBAIAgentClientSnapshot) loadLegacySnapshot(ctx context.Context, current map[string]map[string]string, credentials AWSCredentials) (AIAgentClientSnapshot, bool, error) {
	snapshot, _, err := decodeLegacyDynamoDBAIAgentClientSnapshot(current)
	if err != nil {
		return AIAgentClientSnapshot{}, false, err
	}
	s.partHashes = nil
	return snapshot, true, nil
}

func (s *DynamoDBAIAgentClientSnapshot) querySnapshotItems(ctx context.Context, credentials AWSCredentials) ([]map[string]map[string]string, error) {
	body, err := s.doSnapshotQuery(ctx, credentials)
	if err != nil {
		return nil, err
	}
	var response struct {
		Items []map[string]map[string]string `json:"Items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode DynamoDB AI Agent client snapshot response: %w", err)
	}
	return response.Items, nil
}

func (s *DynamoDBAIAgentClientSnapshot) snapshotLoadTraceAttrs() []TraceAttribute {
	return []TraceAttribute{
		StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), AIAgentClientSnapshotLoad.String()),
	}
}
