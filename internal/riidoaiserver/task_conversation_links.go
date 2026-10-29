package riidoaiserver

// Legacy conversations may share a display identity across Agents. Only
// server-recorded parent links authorize including another participant's history.
func taskConversationLinkedThreadIDs(thread AIAgentTaskThreadRecord, history []AIAgentTaskThreadHistoryRecord) map[string]bool {
	edges := make(map[string][]string)
	for _, row := range history {
		if row.TaskID != thread.TaskID || row.ConversationID != taskThreadConversationID(thread) || row.ParentThreadID == "" {
			continue
		}
		edges[row.ThreadID] = append(edges[row.ThreadID], row.ParentThreadID)
		edges[row.ParentThreadID] = append(edges[row.ParentThreadID], row.ThreadID)
	}
	linked := map[string]bool{thread.ThreadID: true}
	queue := []string{thread.ThreadID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, next := range edges[id] {
			if !linked[next] {
				linked[next] = true
				queue = append(queue, next)
			}
		}
	}
	return linked
}
