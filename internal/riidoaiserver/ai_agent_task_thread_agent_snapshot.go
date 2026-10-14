package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) agentSnapshotFromAgent(agent AgentClientRecord, capturedAt time.Time) *AIAgentTaskThreadAgentSnapshot {
	agentID := agent.AgentID
	if agentID == "" {
		return nil
	}
	return &AIAgentTaskThreadAgentSnapshot{
		AgentID:             agentID,
		WorkspaceID:         s.agentWorkspaceID(agent),
		OwnerPrincipalID:    agent.OwnerPrincipalID,
		Name:                agent.Name,
		ProfileThumbnailURL: agent.ProfileThumbnailURL,
		TmpColor:            agent.TmpColor,
		Visibility:          agent.Visibility,
		RuntimeKind:         agent.RuntimeKind,
		ModelID:             agent.ModelID,
		ModelLabel:          agent.ModelLabel,
		CapturedAt:          capturedAt.UTC(),
	}
}

func copyTaskThreadAgentSnapshot(snapshot *AIAgentTaskThreadAgentSnapshot) *AIAgentTaskThreadAgentSnapshot {
	if snapshot == nil {
		return nil
	}
	copied := *snapshot
	return &copied
}
