package riidoaiserver

import "slices"

func cloneTaskMembership(input map[string][]string) map[string][]string {
	output := make(map[string][]string, len(input))
	for task, ids := range input {
		output[task] = append([]string{}, ids...)
	}
	return output
}
func (s *DevelopmentAIAgentClientStore) ensureTaskMembershipLocked() {
	if s.taskAssignedAgents != nil {
		return
	}
	s.taskAssignedAgents = map[string][]string{}
	for task, threads := range s.taskThreads {
		for _, thread := range threads {
			if taskThreadNeedsRetention(thread) && !slices.Contains(s.taskAssignedAgents[task], thread.AgentID) {
				s.taskAssignedAgents[task] = append(s.taskAssignedAgents[task], thread.AgentID)
			}
		}
	}
}
func (s *DevelopmentAIAgentClientStore) addTaskMemberLocked(task, agent string) {
	s.ensureTaskMembershipLocked()
	if !slices.Contains(s.taskAssignedAgents[task], agent) {
		s.taskAssignedAgents[task] = append(s.taskAssignedAgents[task], agent)
	}
}
func (s *DevelopmentAIAgentClientStore) removeTaskMemberLocked(task, agent string) {
	s.ensureTaskMembershipLocked()
	s.taskAssignedAgents[task] = slices.DeleteFunc(s.taskAssignedAgents[task], func(id string) bool { return id == agent })
}

func taskThreadNeedsRetention(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning, AgentAssignmentStateWaiting:
		return true
	default:
		return false
	}
}
