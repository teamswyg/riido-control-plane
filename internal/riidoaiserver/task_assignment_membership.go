package riidoaiserver

import "sort"

func taskThreadNeedsRetention(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning, AgentAssignmentStateWaiting:
		return true
	default:
		return false
	}
}

func (s *DevelopmentAIAgentClientStore) visibleTaskMemberIDsLocked(principal AuthorizationResult, taskID string) []string {
	members := s.taskAgentMemberships
	if members == nil {
		members = legacyTaskAgentMemberships(s.taskThreads)
	}
	ids := []string{}
	for id, member := range members[taskID] {
		if member.RemovedAt != nil {
			continue
		}
		if _, ok := s.visibleAgent(principal, id); ok {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
