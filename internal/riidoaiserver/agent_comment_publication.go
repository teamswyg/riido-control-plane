package riidoaiserver

import (
	"context"
	"strings"
	"unicode/utf8"
)

func (s *DevelopmentAIAgentClientStore) agentCommentPublication(ctx context.Context, event TaskEvent) (agentCommentPublication, error) {
	if err := ctx.Err(); err != nil {
		return agentCommentPublication{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, exists := s.taskThreadForAssignmentLocked(event.TaskID, event.AgentID, event.AssignmentID)
	if !exists || thread.CommentKind != AgentTaskCommentTaskCompleted || thread.AssignmentState != AgentAssignmentStateCompleted {
		return agentCommentPublication{}, nil
	}
	agent, exists := s.agents[event.AgentID]
	if !exists || !s.aiAgentVisibleTo(AuthorizationResult{WorkspaceID: event.WorkspaceID, PrincipalID: event.CreatedBy}, agent) {
		return agentCommentPublication{}, nil
	}
	message, visible := taskThreadProjectionMessage(thread)
	body := strings.TrimSpace(message.ResultMessage)
	if !visible || body == "" || body == clientMessageTaskCompleted {
		return agentCommentPublication{}, nil
	}
	conversation := taskThreadConversationID(thread)
	root := ""
	if thread.CommentInvocation != nil {
		root = thread.CommentInvocation.RootCommentID
	}
	if root == "" {
		if sourceID, linked := strings.CutPrefix(conversation, "comment:"); linked {
			root = sourceID
		}
	}
	name := strings.TrimSpace(agent.Name)
	if name == "" {
		name = "Riido Agent"
	}
	nameRunes := []rune(name)
	if len(nameRunes) > 120 {
		name = string(nameRunes[:120])
	}
	body = strings.ReplaceAll(body, "\x00", "")
	if len(body) > 32<<10 {
		end := (32 << 10) - len("…")
		for !utf8.RuneStart(body[end]) {
			end--
		}
		body = body[:end] + "…"
	}
	if strings.TrimSpace(body) == "" {
		return agentCommentPublication{}, nil
	}
	return agentCommentPublication{
		WorkspaceID: event.WorkspaceID, TaskID: event.TaskID, RequesterID: event.CreatedBy,
		ConversationID: conversation, ThreadID: thread.ThreadID, MessageID: message.MessageID,
		RootCommentID: root, AgentID: agent.AgentID, AgentName: name, Body: body, Public: agent.Visibility == AgentVisibilityPublic,
	}, nil
}
