package riidoaiserver

import (
	"strconv"
	"strings"
)

// RIID-5842: Riido task lists count only monolith comments. After a task's
// Agent conversation changes, push the absolute number of visible Agent
// conversation messages to the monolith, which adds it to commentsCount.
// Failed pushes are repaired by the next change or history read of the task.

const agentCommentCountPath = "/internal/agent-comment-counts"

// Mirrors the client's visible-message rule (isVisibleThreadMessage):
// progress/tool rows and run-status kinds stay in the activity area.
var agentCommentCountHiddenKinds = map[AgentTaskCommentKind]bool{
	AgentTaskCommentQueuedByBusyAgent: true,
	AgentTaskCommentAssignmentStarted: true,
	AgentTaskCommentRuntimeProgress:   true,
}

// countVisibleAgentCommentMessages counts messages the comments tab shows for the
// task. User messages copied from Riido comments are already monolith comments;
// replies typed in the Agent thread carry a "thread-message:" source and count.
func countVisibleAgentCommentMessages(threads []AIAgentTaskThreadHistoryRecord) int {
	seen := map[string]bool{}
	for _, thread := range threads {
		for index, message := range thread.Messages {
			if message.Role != AIAgentTaskThreadMessageRoleUser && message.Role != AIAgentTaskThreadMessageRoleAgent {
				continue
			}
			if agentCommentCountHiddenKinds[message.CommentKind] {
				continue
			}
			body := message.ResultMessage
			if strings.TrimSpace(body) == "" {
				body = message.Body
			}
			if strings.TrimSpace(body) == "" {
				continue
			}
			source := strings.TrimSpace(message.SourceMessageID)
			if message.Role == AIAgentTaskThreadMessageRoleUser && source != "" && !strings.HasPrefix(source, "thread-message:") {
				continue
			}
			key := message.MessageID
			if key == "" {
				key = source
			}
			if key == "" {
				key = thread.ThreadID + ":" + strconv.Itoa(index)
			}
			seen[key] = true
		}
	}
	return len(seen)
}
