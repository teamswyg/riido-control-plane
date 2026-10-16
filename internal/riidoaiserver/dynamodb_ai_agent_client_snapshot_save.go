package riidoaiserver

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *DynamoDBAIAgentClientSnapshot) save(ctx context.Context, snapshot AIAgentClientSnapshot, credentials AWSCredentials) (err error) {
	startedAt := time.Now()
	var snapshotBytes int64
	var stats dynamoDBAIAgentClientSnapshotWriteStats
	ctx, span := StartTraceSpan(ctx, nil, TraceSpanStart{
		Name: aiAgentClientSnapshotSaveTraceName,
		Kind: TraceSpanKindInternal,
		Attributes: []TraceAttribute{
			StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), AIAgentClientSnapshotSave.String()),
			StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), aiAgentClientSnapshotTraceSurface),
		},
	})
	defer func() {
		span.SetAttributes(aiAgentClientSnapshotSaveTraceAttributes(stats, snapshotBytes)...)
		s.observeAIAgentClientSnapshot(AIAgentClientSnapshotSave, startedAt, snapshotBytes, err)
		FinishTraceSpan(span, err)
	}()
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	if snapshot.SavedAt.IsZero() {
		snapshot.SavedAt = s.now()
	}
	if preserve, _ := ctx.Value(taskAgentPreserveMembershipContextKey{}).(bool); preserve {
		stats, snapshotBytes, err = s.saveMembershipOwnedSnapshot(ctx, snapshot, credentials)
		return err
	}
	items, hashes, encodedBytes, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, s.partHashes)
	if err != nil {
		return err
	}
	snapshotBytes = encodedBytes
	stats = dynamoDBAIAgentClientSnapshotWritePressure(items)
	if err := s.putSnapshotTransaction(ctx, items, snapshot.ExpectedVersion, credentials); err != nil {
		return err
	}
	s.partHashes = hashes
	s.onboardingBaseline = onboardingOwnedSnapshot(snapshot)
	return nil
}
