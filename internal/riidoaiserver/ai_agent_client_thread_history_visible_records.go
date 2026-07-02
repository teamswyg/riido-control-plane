package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) visibleTaskThreadHistoryRecordsLocked(
	principal AuthorizationResult,
	taskID string,
) ([]AIAgentTaskThreadHistoryRecord, map[string]*AIAgentTaskThreadAgentSnapshot) {
	source := s.taskThreads[taskID]
	records := make([]AIAgentTaskThreadHistoryRecord, 0, len(source))
	var snapshots map[string]*AIAgentTaskThreadAgentSnapshot
	streamHref := aiAgentClientEventStreamHref(strings.TrimSpace(principal.WorkspaceID))
	for i := range source {
		s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
		if !s.taskThreadVisibleTo(principal, source[i]) {
			continue
		}
		record := s.taskThreadHistoryRecordLocked(source[i], streamHref)
		records = append(records, record)
		snapshots = appendTaskThreadHistorySnapshot(snapshots, record, source[i])
	}
	s.taskThreads[taskID] = source
	return records, snapshots
}

func appendTaskThreadHistorySnapshot(
	snapshots map[string]*AIAgentTaskThreadAgentSnapshot,
	record AIAgentTaskThreadHistoryRecord,
	thread AIAgentTaskThreadRecord,
) map[string]*AIAgentTaskThreadAgentSnapshot {
	if record.AgentSnapshotID == "" || thread.AgentSnapshot == nil {
		return snapshots
	}
	if snapshots == nil {
		snapshots = map[string]*AIAgentTaskThreadAgentSnapshot{}
	}
	snapshots[record.AgentSnapshotID] = thread.AgentSnapshot
	return snapshots
}
