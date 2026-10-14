package riidoaiserver

func (s *DevelopmentAIAgentClientStore) agentFromTaskThreadLocked(principal AuthorizationResult, thread AIAgentTaskThreadRecord) (AgentClientRecord, bool) {
	s.ensureTaskThreadAgentSnapshotLocked(&thread, thread.StartedAt)
	if !s.taskThreadVisibleTo(principal, thread) || thread.AgentSnapshot == nil {
		return AgentClientRecord{}, false
	}
	agent := agentRecordFromTaskThreadSnapshot(thread)
	agent.AssignedTaskCount = s.activeTaskThreadCountForAgentLocked(agent.AgentID)
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
	return agent, true
}

func agentRecordFromTaskThreadSnapshot(thread AIAgentTaskThreadRecord) AgentClientRecord {
	snapshot := thread.AgentSnapshot
	return AgentClientRecord{
		AgentID:             snapshot.AgentID,
		OwnerPrincipalID:    snapshot.OwnerPrincipalID,
		WorkspaceID:         snapshot.WorkspaceID,
		Name:                snapshot.Name,
		ProfileThumbnailURL: snapshot.ProfileThumbnailURL,
		TmpColor:            snapshot.TmpColor,
		Visibility:          snapshot.Visibility,
		RuntimeKind:         snapshot.RuntimeKind,
		ModelID:             snapshot.ModelID,
		ModelLabel:          snapshot.ModelLabel,
		WorkStatus:          thread.WorkStatus,
		CreatedAt:           snapshot.CapturedAt,
		UpdatedAt:           snapshot.CapturedAt,
	}
}

func (s *DevelopmentAIAgentClientStore) activeTaskThreadCountForAgentLocked(agentID string) int {
	count := 0
	for _, threads := range s.taskThreads {
		for _, thread := range threads {
			if thread.AgentID == agentID && taskThreadHasActiveStream(thread) {
				count++
			}
		}
	}
	return count
}
