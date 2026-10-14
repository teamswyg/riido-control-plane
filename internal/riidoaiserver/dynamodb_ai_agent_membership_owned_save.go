package riidoaiserver

func (s *DynamoDBAIAgentClientSnapshot) requiresFreshTaskAgentMembership() bool { return true }

func sameTaskAgentMembershipBaseline(a, b AIAgentTaskMembership) bool {
	if a.DurableGeneration != b.DurableGeneration || a.IntentGateAssignmentID != b.IntentGateAssignmentID {
		return false
	}
	if a.RemovedAt == nil || b.RemovedAt == nil {
		return a.RemovedAt == nil && b.RemovedAt == nil
	}
	return a.RemovedAt.Equal(*b.RemovedAt)
}
