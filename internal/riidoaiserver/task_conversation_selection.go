package riidoaiserver

func selectAIAgentHistoryThreadForFollowup(history AIAgentTaskThreadHistoryCollectionResponse, taskID, threadID string) (AIAgentTaskThreadRecord, error) {
	if history.TaskID != taskID {
		return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
	}
	for _, thread := range history.Threads {
		if thread.ThreadID != threadID || thread.TaskID != taskID {
			continue
		}
		return AIAgentTaskThreadRecord{
			ThreadID: thread.ThreadID, ConversationID: thread.ConversationID,
			ParentThreadID: thread.ParentThreadID, TaskID: thread.TaskID,
			AssignmentID: thread.AssignmentID, AgentID: thread.AgentID,
			RunID: thread.RunID, WorkStatus: thread.WorkStatus,
			AssignmentState: thread.AssignmentState,
		}, nil
	}
	return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
}
