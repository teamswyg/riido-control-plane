package riidoaiserver

const (
	aiAgentClientSnapshotSaveTraceName = "ai_agent_client.snapshot_save"
	aiAgentClientSnapshotTraceSurface  = "ai_agent_client_snapshot"

	riidoSnapshotBytesEncodedKey = "riido.snapshot.bytes_encoded"
	riidoSnapshotItemsWrittenKey = "riido.snapshot.items_written"
	riidoSnapshotPartsSkippedKey = "riido.snapshot.parts_skipped"
	riidoSnapshotPartsWrittenKey = "riido.snapshot.parts_written"
)

type dynamoDBAIAgentClientSnapshotWriteStats struct {
	itemsWritten int
	partsSkipped int
	partsWritten int
}

func aiAgentClientSnapshotSaveTraceAttributes(stats dynamoDBAIAgentClientSnapshotWriteStats, bytes int64) []TraceAttribute {
	return []TraceAttribute{
		Int64TraceAttribute(riidoSnapshotBytesEncodedKey, bytes),
		Int64TraceAttribute(riidoSnapshotItemsWrittenKey, int64(stats.itemsWritten)),
		Int64TraceAttribute(riidoSnapshotPartsSkippedKey, int64(stats.partsSkipped)),
		Int64TraceAttribute(riidoSnapshotPartsWrittenKey, int64(stats.partsWritten)),
	}
}
