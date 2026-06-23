package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) appendThreadUserMessageLocked(response AIAgentTaskActionResponse, body, sourceMessageID string) {
	body = strings.TrimSpace(body)
	if body == "" || strings.TrimSpace(response.ThreadID) == "" {
		return
	}
	now := time.Now().UTC()
	message := AIAgentTaskThreadHistoryMessage{
		MessageID:       taskThreadUserMessageID(response.ThreadID, response.AssignmentID, sourceMessageID, body, now),
		Role:            AIAgentTaskThreadMessageRoleUser,
		AssignmentID:    strings.TrimSpace(response.AssignmentID),
		RunID:           strings.TrimSpace(response.RunID),
		SourceMessageID: strings.TrimSpace(sourceMessageID),
		Body:            body,
		ObservedAt:      now,
	}
	s.appendThreadHistoryMessageLocked(response.ThreadID, message)
}

func (s *DevelopmentAIAgentClientStore) appendThreadHistoryMessageLocked(threadID string, message AIAgentTaskThreadHistoryMessage) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" || strings.TrimSpace(message.MessageID) == "" {
		return
	}
	if s.taskThreadMessages == nil {
		s.taskThreadMessages = map[string][]AIAgentTaskThreadHistoryMessage{}
	}
	source := s.taskThreadMessages[threadID]
	for _, existing := range source {
		if existing.MessageID == message.MessageID {
			return
		}
	}
	s.taskThreadMessages[threadID] = append(source, message)
}
