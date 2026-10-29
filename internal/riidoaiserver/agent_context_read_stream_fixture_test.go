package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type contextReadStreamStore struct {
	*DevelopmentAIAgentClientStore
	history []ClientStreamEvent
	live    <-chan ClientStreamEvent
}

func (s *contextReadStreamStore) SubscribeAIAgentClientEvents(context.Context, AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	return s.history, s.live, func() {}, nil
}

func contextReadStreamServer(t *testing.T) (Server, *runReadsTestContext, *runReadsTestAuthorizer, ClientStreamEvent) {
	t.Helper()
	s, reader, auth := runReadsTestServer(t, nil)
	store := s.aiAgent.(*DevelopmentAIAgentClientStore)
	agent := store.agents["agent-owned-codex"]
	agent.Visibility = AgentVisibilityPublic
	store.agents[agent.AgentID] = agent
	event := ClientStreamEvent{Seq: 1, EventType: AgentClientEventContextRead, Payload: AgentContextReadEvent{
		EventType: AgentClientEventContextRead, TaskID: "task", ThreadID: "thread", RunID: "run", AssignmentID: "assignment", AgentID: agent.AgentID,
		Sources:        []AgentContextReadSource{{Kind: "task", ID: "task", Part: "body", Status: "served"}, {Kind: "link", ID: "hidden-resource", Part: "metadata", Version: "hidden-hash", Status: "served"}},
		OmittedSources: 17,
	}}
	return s, reader, auth, event
}

func contextReadStreamRequest(s Server, mode string, events ...ClientStreamEvent) *httptest.ResponseRecorder {
	store := &contextReadStreamStore{DevelopmentAIAgentClientStore: s.aiAgent.(*DevelopmentAIAgentClientStore)}
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/events"
	if mode == "replay" {
		store.history = events
		path += "?replay=1"
	} else {
		live := make(chan ClientStreamEvent, len(events))
		for _, event := range events {
			live <- event
		}
		close(live)
		store.live = live
	}
	s.aiAgent = store
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Header.Set(aiAgentTokenHeader, "viewer-token")
	r.Header.Set("Authorization", "Bearer unrelated-token")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
