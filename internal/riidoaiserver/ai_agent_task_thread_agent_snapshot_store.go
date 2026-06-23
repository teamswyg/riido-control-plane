package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) snapshotForTaskThreadLocked(agentID string, capturedAt time.Time) *AIAgentTaskThreadAgentSnapshot {
	agent, ok := s.agents[agentID]
	if !ok {
		return nil
	}
	return s.agentSnapshotFromAgent(agent, capturedAt)
}

func (s *DevelopmentAIAgentClientStore) ensureTaskThreadAgentSnapshotLocked(thread *AIAgentTaskThreadRecord, capturedAt time.Time) {
	if thread == nil || thread.AgentSnapshot != nil {
		return
	}
	thread.AgentSnapshot = s.snapshotForTaskThreadLocked(thread.AgentID, capturedAt)
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
			}
		}
		s.taskThreads[taskID] = threads
	}
}
