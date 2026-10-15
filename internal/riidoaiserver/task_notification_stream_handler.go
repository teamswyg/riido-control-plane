package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
)

// Multiplex workspace subscriptions into one browser connection. HTTP/1 clients
// otherwise exhaust their per-origin connection pool before loading task data.
func (s Server) handleTaskNotificationStream(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["workspace_id"]
	if len(ids) == 0 || len(ids) > 100 {
		writeError(w, http.StatusBadRequest, "between 1 and 100 workspaces required")
		return
	}
	subscriber, ok := s.aiAgent.(AIAgentClientEventSubscriber)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "live events unavailable")
		return
	}
	principals := make(map[string]AuthorizationResult)
	owner := ""
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			writeError(w, http.StatusBadRequest, "workspace required")
			return
		}
		if _, exists := principals[id]; exists {
			continue
		}
		scoped := requestWithAIAgentWorkspaceIDAndPath(r, id, r.URL.Path)
		principal, allowed := s.authorizeAIAgentClient(w, scoped, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStream})
		if !allowed {
			return
		}
		if principal.WorkspaceID != id || (owner != "" && principal.PrincipalID != owner) {
			writeError(w, http.StatusForbidden, "inconsistent notification scope")
			return
		}
		owner = principal.PrincipalID
		principals[id] = principal
	}
	ctx, stop := context.WithCancel(r.Context())
	defer stop()
	merged := make(chan ClientStreamEvent, 32)
	for _, principal := range principals {
		_, live, cancel, err := subscriber.SubscribeAIAgentClientEvents(ctx, principal)
		if cancel != nil {
			defer cancel()
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case event, open := <-live:
					if !open {
						stop()
						return
					}
					if event.EventType != AgentClientEventTaskNotification {
						continue
					}
					select {
					case merged <- event:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	// Never replay notifications. Membership and task access are checked again for
	// every live delivery, including after a long-running stream loses access.
	serveAIAgentClientEventStream(w, r.WithContext(ctx), nil, merged, func(event ClientStreamEvent) bool {
		notification, valid := event.Payload.(AgentTaskNotificationEvent)
		if !valid {
			return false
		}
		principal, allowed := principals[notification.WorkspaceID]
		return allowed && s.clientNotificationAuthorized(r, principal, event)
	})
}
