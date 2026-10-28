package riidoaiserver

import (
	"context"
	"strings"
)

// The root is supplied only by the authenticated comment evidence service.
// A triggering reply is a message identity, never a new conversation identity.
func (s Server) prepareCommentConversation(ctx context.Context, principal AuthorizationResult, evidence verifiedCommentDelivery, publication *AssignAIAgentTaskRequest, req *AssignRequest, owner string) error {
	root := strings.TrimSpace(evidence.RootCommentID)
	if root == "" {
		root = evidence.CommentID
	} // Rolling compatibility with old evidence services.
	publication.conversationID = "comment:" + root
	publication.sourceMessageObservedAt = evidence.CommentCreatedAt
	for _, message := range evidence.ConversationMessages {
		if message.ID == evidence.CommentID && publication.sourceMessageObservedAt.IsZero() {
			publication.sourceMessageObservedAt = message.CreatedAt
			break
		}
	}
	history, err := s.commentConversationHistory(ctx, principal, evidence, publication)
	if err != nil {
		return err
	}
	var previous *AIAgentTaskThreadHistoryRecord
	for i := range history.Threads {
		row := &history.Threads[i]
		if row.TaskID != evidence.TaskID || row.ConversationID != publication.conversationID {
			continue
		}
		if publication.parentThreadID == "" {
			publication.parentThreadID = row.ThreadID
		}
		if row.AgentID == evidence.AgentID && (previous == nil || row.StartedAt.After(previous.StartedAt)) {
			previous = row
		}
	}
	selected := AIAgentTaskThreadRecord{TaskID: evidence.TaskID, AgentID: evidence.AgentID, ConversationID: publication.conversationID}
	if previous != nil {
		selected, err = selectAIAgentHistoryThreadForFollowup(history, evidence.TaskID, previous.ThreadID)
		if err != nil {
			return err
		}
		if !taskThreadMessageStartsNewExecution(selected) {
			return ErrAIAgentTaskThreadConflict
		}
		if err := s.applyThreadNativeResume(ctx, principal, selected, req, owner); err != nil {
			return err
		}
		if nativeFollowupProvider(req.RuntimeProvider) && !isNonDurableAssignmentID(selected.AssignmentID) && req.ResumeSessionID == "" {
			return ErrAIAgentTaskThreadConflict
		}
		publication.parentThreadID = previous.ThreadID
	}
	var contextText strings.Builder
	contextHistory := appendVerifiedCommentContext(history.Threads, evidence, selected)
	contextSourceID := evidence.CommentID
	if publication.skipSourceMessage {
		contextSourceID = ""
	}
	appendTaskConversationContext(&contextText, selected, CreateAIAgentTaskThreadMessageRequest{SourceMessageID: contextSourceID}, contextHistory)
	req.Prompt += contextText.String()
	return nil
}
