package riidoaiserver

import (
	"context"
	"crypto/rand"
	"strings"
)

type preparedThreadReply struct {
	Message  CreateAIAgentTaskThreadMessageRequest
	Prepared preparedThreadMessage
}

func (s Server) prepareThreadReplies(ctx context.Context, principal AuthorizationResult, bearer, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) ([]preparedThreadReply, error) {
	ids, err := commentMentionIDs(req.Body)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		ids = []string{""}
	} else {
		if req.RequestID == "" {
			req.RequestID = rand.Text()
		}
		if req.SourceMessageID == "" {
			// One human reply remains one message even when it names several Agents.
			req.SourceMessageID = "thread-message:" + taskRequestReceiptKey(principal, taskID, AssignAIAgentTaskRequest{RequestID: threadMessageRequestID(threadID, req.RequestID)})
		}
	}
	var replies []preparedThreadReply
	for _, id := range ids {
		req.targetAgentID = id
		prepared, err := s.prepareThreadMessage(ctx, principal, bearer, taskID, threadID, req)
		if err != nil {
			return nil, err
		}
		replies = append(replies, preparedThreadReply{Message: req, Prepared: prepared})
	}
	return replies, nil
}

func threadMessageExecutionThread(selected AIAgentTaskThreadRecord, history AIAgentTaskThreadHistoryCollectionResponse, targetAgentID string) AIAgentTaskThreadRecord {
	if targetAgentID == "" || targetAgentID == selected.AgentID {
		return selected
	}
	var latest *AIAgentTaskThreadHistoryRecord
	conversationID := taskThreadConversationID(selected)
	linked := taskConversationLinkedThreadIDs(selected, history.Threads)
	for i := range history.Threads {
		row := &history.Threads[i]
		if !strings.HasPrefix(conversationID, "comment:") && !linked[row.ThreadID] {
			continue
		}
		if row.TaskID == selected.TaskID && row.AgentID == targetAgentID && row.ConversationID == conversationID && (latest == nil || row.StartedAt.After(latest.StartedAt)) {
			latest = row
		}
	}
	if latest != nil {
		target, _ := selectAIAgentHistoryThreadForFollowup(history, selected.TaskID, latest.ThreadID)
		return target
	}
	// A new participant receives conversation context, never another Agent's session.
	return AIAgentTaskThreadRecord{TaskID: selected.TaskID, AgentID: targetAgentID, ConversationID: conversationID}
}
