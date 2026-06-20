package riidoaiserver

import (
	"context"
	"slices"
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
		for _, thread := range slices.Backward(threads) {
			if !taskThreadHasActiveStream(thread) {
				continue
			}
			agent, ok := s.agents[thread.AgentID]
			if !ok || !s.aiAgentVisibleTo(principal, agent) {
				continue
			}
			componentID := strings.TrimSpace(thread.TaskID)
			if componentID == "" {
				componentID = strings.TrimSpace(taskID)
			}
			if componentID == "" {
				continue
			}
			profiles[componentID] = assignedAgentProfileFromAgent(agent)
			break
		}
	}
	return AssignedAgentProfileMapResponse{
		SchemaVersion:         SchemaVersion,
		WorkspaceID:           s.workspaceScope(principal),
		AssignedAgentProfiles: profiles,
	}, nil
}
