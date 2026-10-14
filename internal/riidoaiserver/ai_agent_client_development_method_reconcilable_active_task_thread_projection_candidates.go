package riidoaiserver

import (
	"sort"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) reconcilableActiveTaskThreadProjectionCandidates(principal AuthorizationResult, taskID string) []AIAgentTaskThreadRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	taskIDs := []string{}
	if taskID != "" {
		taskIDs = []string{taskID}
	} else {
		for id := range s.taskThreads {
			taskIDs = append(taskIDs, id)
		}
		sort.Strings(taskIDs)
	}
	out := []AIAgentTaskThreadRecord{}
	for _, id := range taskIDs {
		for _, thread := range s.taskThreads[id] {
			if !taskThreadHasActiveStream(thread) || strings.TrimSpace(thread.AssignmentID) == "" {
				continue
			}
			agent, ok := s.agents[thread.AgentID]
			if !ok || !s.aiAgentVisibleTo(principal, agent) {
				continue
			}
			out = append(out, copyTaskThread(thread))
		}
	}
	return out
}
