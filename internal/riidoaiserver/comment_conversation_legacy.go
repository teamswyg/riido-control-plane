package riidoaiserver

import (
	"context"
	"slices"
	"strings"
)

// Adoption is lazy and uses only Monolith-verified membership in this root.
// Legacy execution IDs remain unchanged, including their native session pins.
func (s Server) commentConversationHistory(ctx context.Context, principal AuthorizationResult, evidence verifiedCommentDelivery, publication *AssignAIAgentTaskRequest) (AIAgentTaskThreadHistoryCollectionResponse, error) {
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, evidence.TaskID)
	if err != nil {
		return history, err
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, evidence.TaskID)
	if err != nil {
		return history, err
	}
	sources := map[string]bool{evidence.RootCommentID: true}
	for _, id := range evidence.ConversationCommentIDs {
		sources[id] = true
	}
	for _, message := range evidence.ConversationMessages {
		sources[message.ID] = true
	}
	delete(sources, "")
	for _, thread := range threads.Threads {
		conversation := taskThreadConversationID(thread)
		if thread.TaskID == evidence.TaskID && sources[thread.SourceCommentID] && !strings.HasPrefix(conversation, "comment:") && !slices.Contains(publication.legacyCommentConversations, conversation) {
			publication.legacyCommentConversations = append(publication.legacyCommentConversations, conversation)
		}
	}
	for i := range history.Threads {
		if slices.Contains(publication.legacyCommentConversations, history.Threads[i].ConversationID) {
			history.Threads[i].ConversationID = publication.conversationID
		}
	}
	return history, nil
}

func (s *DevelopmentAIAgentClientStore) adoptLegacyCommentConversationsLocked(taskID string, req AssignAIAgentTaskRequest) {
	for i := range s.taskThreads[taskID] {
		row := &s.taskThreads[taskID][i]
		if !strings.HasPrefix(taskThreadConversationID(*row), "comment:") && slices.Contains(req.legacyCommentConversations, taskThreadConversationID(*row)) {
			row.ConversationID = req.conversationID
			s.dropTaskThreadHistoryCacheLocked(row.ThreadID)
		}
	}
}
