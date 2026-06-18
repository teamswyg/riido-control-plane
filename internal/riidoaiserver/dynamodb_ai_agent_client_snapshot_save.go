package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

func (s *DynamoDBAIAgentClientSnapshot) save(ctx context.Context, snapshot AIAgentClientSnapshot, credentials AWSCredentials) (err error) {
	startedAt := time.Now()
	var snapshotBytes int64
	defer func() {
		s.observeAIAgentClientSnapshot(AIAgentClientSnapshotSave, startedAt, snapshotBytes, err)
	}()
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	if snapshot.SavedAt.IsZero() {
		snapshot.SavedAt = s.now()
	}
	items, hashes, encodedBytes, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, s.partHashes)
	if err != nil {
		return err
	}
	snapshotBytes = encodedBytes
	if err := s.putSnapshotItems(ctx, items, credentials); err != nil {
		return err
	}
	s.partHashes = hashes
	return nil
}

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotItems(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) error {
	for _, item := range items {
		if err := s.putSnapshotItem(ctx, item, credentials); err != nil {
			return err
		}
	}
	return nil
}
