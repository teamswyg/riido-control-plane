package riidoaiserver

// This private observation fences membership removal, not execution capacity.
// An empty latestThreadID explicitly observes the absence of retained history.
type taskAgentMembershipRemovalFence struct {
	taskID                 string
	agentID                string
	latestThreadID         string
	durableGeneration      string
	intentGateAssignmentID string
}

func observeTaskAgentMembershipRemoval(taskID, agentID string, threads []AIAgentTaskThreadRecord) *taskAgentMembershipRemovalFence {
	fence := &taskAgentMembershipRemovalFence{taskID: taskID, agentID: agentID}
	for i := len(threads) - 1; i >= 0; i-- {
		if threads[i].AgentID == agentID {
			fence.latestThreadID = threads[i].ThreadID
			break
		}
	}
	return fence
}

func validateTaskAgentMembershipRemovalObservation(taskID, agentID string, expected *taskAgentMembershipRemovalFence, threads []AIAgentTaskThreadRecord) error {
	if expected == nil {
		return nil
	}
	current := observeTaskAgentMembershipRemoval(taskID, agentID, threads)
	current.durableGeneration = expected.durableGeneration
	current.intentGateAssignmentID = expected.intentGateAssignmentID
	if *current != *expected {
		return ErrAIAgentMembershipChanged
	}
	return nil
}
