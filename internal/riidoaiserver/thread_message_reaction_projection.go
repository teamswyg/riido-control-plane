package riidoaiserver

// The durable execution projection owns status. Deriving its reaction on reads
// keeps retries and snapshot repair consistent without a second lifecycle writer.
func agentStatusReaction(thread AIAgentTaskThreadRecord) (string, bool) {
	switch string(thread.AssignmentState) {
	case "cancelling", "cancelled", "stopping", "stopped":
		return "", true
	}
	switch thread.WorkStatus {
	case AgentWorkStatusCompleted:
		return "2705", true
	case AgentWorkStatusFailed:
		return "274c", true
	case AgentWorkStatusQueued, AgentWorkStatusRunning, AgentWorkStatusWaitingForUser:
		return "1f440", true
	}
	return "", false
}

func (s *DevelopmentAIAgentClientStore) projectAgentMessageReactionsLocked(records []AIAgentTaskThreadHistoryRecord) {
	for _, record := range records {
		for _, run := range s.taskThreads[record.TaskID] {
			sourceID := run.SourceMessageID
			if run.ReactionSourceMessageID != nil {
				sourceID = *run.ReactionSourceMessageID
			}
			if run.ThreadID != record.ThreadID || sourceID == "" {
				continue
			}
			emoji, known := agentStatusReaction(run)
			if !known {
				continue
			}
			targetThread := run.ThreadID
			if call := run.CommentInvocation; call != nil && call.OriginThreadID != "" {
				targetThread = call.OriginThreadID
			}
			for i := range records {
				if records[i].ThreadID != targetThread {
					continue
				}
				for j := range records[i].Messages {
					message := &records[i].Messages[j]
					if message.Role != AIAgentTaskThreadMessageRoleUser && message.Role != AIAgentTaskThreadMessageRoleAgent {
						continue
					}
					if message.MessageID != sourceID && message.SourceMessageID != sourceID {
						continue
					}
					actor := AIAgentMessageReactionActor{Kind: "agent", ID: run.AgentID}
					for _, status := range []string{"1f440", "2705", "274c"} {
						message.Reactions = mutateMessageReaction(message.Reactions, actor, status, true)
					}
					if emoji != "" {
						message.Reactions = mutateMessageReaction(message.Reactions, actor, emoji, false)
					}
				}
			}
		}
	}
}
