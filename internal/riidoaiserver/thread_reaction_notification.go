package riidoaiserver

import (
	"context"
)

func (s Server) notifyMemberMessageReaction(ctx context.Context, principal AuthorizationResult, response AIAgentTaskThreadMessageReactionsResponse, emoji string, remove bool) {
	if remove {
		emoji = memberReactionEmoji(response.Reactions, principal.PrincipalID)
	}
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, response.TaskID)
	if err != nil {
		return
	}
	for _, thread := range history.Threads {
		if thread.ThreadID != response.ThreadID {
			continue
		}
		for _, message := range thread.Messages {
			if message.MessageID != response.MessageID || message.Role != AIAgentTaskThreadMessageRoleUser || message.AuthorPrincipalID == "" || message.AuthorPrincipalID == principal.PrincipalID {
				continue
			}
			body := map[string]string{"workspace_id": principal.WorkspaceID, "task_id": response.TaskID, "thread_id": response.ThreadID, "message_id": response.MessageID, "recipient_principal_id": message.AuthorPrincipalID, "actor_principal_id": principal.PrincipalID, "emoji": emoji, "excerpt": message.Body}
			s.queueMemberReactionNotification(memberReactionNotification{key: response.ThreadID + "/" + response.MessageID + "/" + principal.PrincipalID, revision: response.revision, body: body})
		}
	}
}

// After a removal the member's remaining reaction, if any, stays notified.
func memberReactionEmoji(reactions []AIAgentMessageReaction, principalID string) string {
	for _, reaction := range reactions {
		for _, actor := range reaction.Actors {
			if actor.Kind == "member" && actor.ID == principalID {
				return reaction.Emoji
			}
		}
	}
	return ""
}
