package riidoaiserver

import "net/http"

func (s Server) handleAIAgentClientEvents(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStream})
	if !ok {
		return
	}
	var (
		events []ClientStreamEvent
		live   <-chan ClientStreamEvent
		cancel func()
		err    error
	)
	if subscriber, ok := s.aiAgent.(AIAgentClientEventSubscriber); ok {
		events, live, cancel, err = subscriber.SubscribeAIAgentClientEvents(r.Context(), principal)
		if cancel != nil {
			defer cancel()
		}
	} else {
		events, err = s.aiAgent.AIAgentClientEvents(r.Context(), principal)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	serveAIAgentClientEventStream(w, r, events, live)
}
