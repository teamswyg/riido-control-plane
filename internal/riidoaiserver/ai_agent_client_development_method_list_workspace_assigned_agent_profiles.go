package riidoaiserver

import (
	"context"
	"sort"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ListWorkspaceAssignedAgentProfiles(ctx context.Context, principal AuthorizationResult) (AssignedAgentProfileMapResponse, error) {
	if err := ctx.Err(); err != nil {
		return AssignedAgentProfileMapResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles := make(map[string]AssignedAgentProfile)
	taskIDs := make([]string, 0, len(s.taskThreads))
	for taskID := range s.taskThreads {
		taskIDs = append(taskIDs, taskID)
	}
	sort.Strings(taskIDs)
	for _, taskID := range taskIDs {
		threads := s.taskThreads[taskID]
		for i := len(threads) - 1; i >= 0; i-- {
			thread := threads[i]
			if !taskThreadHasActiveStream(thread) {
				continue
			}
			s.ensureTaskThreadAgentSnapshotLocked(&threads[i], threads[i].StartedAt)
			thread = threads[i]
			if !s.taskThreadVisibleTo(principal, thread) {
				continue
			}
			componentID := strings.TrimSpace(thread.TaskID)
			if componentID == "" {
				componentID = strings.TrimSpace(taskID)
			}
			if componentID == "" {
				continue
			}
			agent, ok := s.agents[thread.AgentID]
			if ok && s.aiAgentVisibleTo(principal, agent) {
				profiles[componentID] = assignedAgentProfileFromAgent(agent)
				break
			}
			if profile, ok := assignedAgentProfileFromSnapshot(thread.AgentSnapshot); ok {
				profiles[componentID] = profile
				break
			}
		}
		s.taskThreads[taskID] = threads
	}
	return AssignedAgentProfileMapResponse{
		SchemaVersion:         SchemaVersion,
		WorkspaceID:           s.workspaceScope(principal),
		AssignedAgentProfiles: profiles,
	}, nil
}
