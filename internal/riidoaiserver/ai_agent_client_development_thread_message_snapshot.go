package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) updateTaskThreadMessageAgentSnapshotLocked(
	thread *AIAgentTaskThreadRecord,
	response AIAgentTaskActionResponse,
	now time.Time,
) {
	if thread.AgentSnapshot != nil {
		return
	}
	thread.AgentSnapshot = copyTaskThreadAgentSnapshot(response.AgentSnapshot)
	s.ensureTaskThreadAgentSnapshotLocked(thread, now)
}
