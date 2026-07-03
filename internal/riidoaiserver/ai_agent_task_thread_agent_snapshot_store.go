package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) snapshotForTaskThreadLocked(agentID string, capturedAt time.Time) *AIAgentTaskThreadAgentSnapshot {
	agent, ok := s.agents[agentID]
	if !ok {
		return s.deletedAgentSnapshot(agentID, capturedAt)
	}
	return s.agentSnapshotFromAgent(agent, capturedAt)
}

func (s *DevelopmentAIAgentClientStore) deletedAgentSnapshot(agentID string, capturedAt time.Time) *AIAgentTaskThreadAgentSnapshot {
	if agentID == "" {
		return nil
	}
	if capturedAt.IsZero() {
		capturedAt = time.Now().UTC()
	}
	return &AIAgentTaskThreadAgentSnapshot{
		AgentID:     agentID,
		WorkspaceID: s.workspaceScope(AuthorizationResult{}),
		Name:        "삭제된 에이전트",
		TmpColor:    "#94A3B8",
		Visibility:  AgentVisibilityPublic,
		CapturedAt:  capturedAt.UTC(),
	}
}

func (s *DevelopmentAIAgentClientStore) ensureTaskThreadAgentSnapshotLocked(thread *AIAgentTaskThreadRecord, capturedAt time.Time) {
	if thread == nil {
		return
	}
	if thread.AgentSnapshot == nil {
		thread.AgentSnapshot = s.snapshotForTaskThreadLocked(thread.AgentID, capturedAt)
	}
	if thread.AgentSnapshotID == "" {
		thread.AgentSnapshotID = taskThreadAgentSnapshotID(thread.AgentSnapshot)
	}
}

func (s *DevelopmentAIAgentClientStore) snapshotAgentTaskThreadsLocked(agent AgentClientRecord, capturedAt time.Time) {
	snapshot := s.agentSnapshotFromAgent(agent, capturedAt)
	if snapshot == nil {
		return
	}
	for taskID, threads := range s.taskThreads {
		for i := range threads {
			if threads[i].AgentID == agent.AgentID && threads[i].AgentSnapshot == nil {
				threads[i].AgentSnapshot = copyTaskThreadAgentSnapshot(snapshot)
				threads[i].AgentSnapshotID = ""
				s.ensureTaskThreadAgentSnapshotLocked(&threads[i], capturedAt)
			}
		}
		s.taskThreads[taskID] = threads
	}
}
