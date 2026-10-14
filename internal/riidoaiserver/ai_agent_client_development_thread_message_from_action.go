package riidoaiserver

import (
	"strings"
	"time"
)

func taskThreadMessageFromAction(response AIAgentTaskActionResponse, sourceMessageID, conversationID, parentThreadID string, now time.Time) AIAgentTaskThreadRecord {
	conversationID = defaultTaskThreadConversationID(conversationID, response.ThreadID)
	return AIAgentTaskThreadRecord{
		ThreadID:           response.ThreadID,
		ConversationID:     conversationID,
		ParentThreadID:     strings.TrimSpace(parentThreadID),
		TaskID:             response.TaskID,
		AssignmentID:       response.AssignmentID,
		AgentID:            response.AgentID,
		AgentSnapshot:      copyTaskThreadAgentSnapshot(response.AgentSnapshot),
		RunID:              response.RunID,
		SourceMessageID:    strings.TrimSpace(sourceMessageID),
		WorkStatus:         response.WorkStatus,
		AssignmentState:    response.AssignmentState,
		CommentKind:        response.CommentKind,
		Message:            response.Message,
		ResultMessage:      response.ResultMessage,
		FailureDiagnostics: copyFailureDiagnostics(response.FailureDiagnostics),
		StartedAt:          now,
		Lines:              []AgentThreadProgressLine{},
	}
}

func defaultTaskThreadConversationID(conversationID, threadID string) string {
	if id := strings.TrimSpace(conversationID); id != "" {
		return id
	}
	return strings.TrimSpace(threadID)
}
