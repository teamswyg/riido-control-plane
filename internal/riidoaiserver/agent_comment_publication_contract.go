package riidoaiserver

import "context"

type agentCommentPublication struct {
	WorkspaceID    string `json:"workspace_id"`
	TaskID         string `json:"task_id"`
	RequesterID    string `json:"requester_id"`
	ConversationID string `json:"conversation_id"`
	ThreadID       string `json:"thread_id"`
	MessageID      string `json:"message_id"`
	RootCommentID  string `json:"root_comment_id"`
	AgentID        string `json:"agent_id"`
	AgentName      string `json:"agent_name"`
	Body           string `json:"body"`
	Public         bool   `json:"public"`
}

type agentCommentPublicationReader interface {
	agentCommentPublication(context.Context, TaskEvent) (agentCommentPublication, error)
}

func (s *PersistentAIAgentClientStore) agentCommentPublication(ctx context.Context, event TaskEvent) (agentCommentPublication, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return agentCommentPublication{}, err
	}
	return s.DevelopmentAIAgentClientStore.agentCommentPublication(ctx, event)
}
