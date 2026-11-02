package riidoaiserver

import (
	"net/http"
)

// Recheck task access at delivery time. Workspace membership or access to the
// shared agent alone is insufficient; revoked task access must fail closed.
func (s Server) clientNotificationAuthorized(r *http.Request, principal AuthorizationResult, event ClientStreamEvent) bool {
	if event.EventType != AgentClientEventTaskNotification {
		return true
	}
	notification, ok := event.Payload.(AgentTaskNotificationEvent)
	if !ok || notification.RecipientPrincipalID != principal.PrincipalID || notification.WorkspaceID != principal.WorkspaceID || s.config.Authorizer == nil {
		return false
	}
	token, ok := requestToken(r)
	if !ok {
		return false
	}
	current, err := s.config.Authorizer.Authorize(r.Context(), token, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead,
		WorkspaceID: notification.WorkspaceID, TaskID: notification.TaskID,
	})
	return err == nil && current.PrincipalID == principal.PrincipalID && (current.WorkspaceID == "" || current.WorkspaceID == notification.WorkspaceID)
}
