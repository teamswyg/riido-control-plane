package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIAgentClientReplayEventsAfterLastEventID(t *testing.T) {
	events := []ClientStreamEvent{{Seq: 41}, {Seq: 42}, {Seq: 43}}

	got := aiAgentClientReplayEventsAfterLastEventID(events, "42")

	if len(got) != 1 || got[0].Seq != 43 {
		t.Fatalf("replay events = %+v, want only seq 43", got)
	}
}

func TestAIAgentClientReplayEventsIgnoreInvalidLastEventID(t *testing.T) {
	events := []ClientStreamEvent{{Seq: 41}, {Seq: 42}}

	for _, cursor := range []string{"", "not-a-sequence", "-1"} {
		got := aiAgentClientReplayEventsAfterLastEventID(events, cursor)
		if len(got) != len(events) {
			t.Fatalf("cursor %q replay count = %d, want %d", cursor, len(got), len(events))
		}
	}
}

func TestAIAgentClientReplayEventsAfterLatestIDReturnsEmpty(t *testing.T) {
	events := []ClientStreamEvent{{Seq: 41}, {Seq: 42}}

	got := aiAgentClientReplayEventsAfterLastEventID(events, "42")

	if len(got) != 0 {
		t.Fatalf("replay events = %+v, want none", got)
	}
}

func TestHTTPAIAgentClientEventsHonorsLastEventID(t *testing.T) {
	store := staticClientEventsStore{
		events: []ClientStreamEvent{
			{Seq: 41, EventType: "test", Payload: map[string]string{"value": "old-41"}},
			{Seq: 42, EventType: "test", Payload: map[string]string{"value": "old-42"}},
			{Seq: 43, EventType: "test", Payload: map[string]string{"value": "new-43"}},
		},
	}
	server := newClientEventsErrorTestServer(t, nil, store)
	req := httptest.NewRequest(http.MethodGet, clientEventsErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Last-Event-ID", "42")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	body := resp.Body.String()
	if resp.Code != http.StatusOK || strings.Contains(body, "old-41") ||
		strings.Contains(body, "old-42") || !strings.Contains(body, "id: 43\n") {
		t.Fatalf("events status/body = %d/%q, want only event 43", resp.Code, body)
	}
}

type staticClientEventsStore struct {
	AIAgentClientStore
	events []ClientStreamEvent
}

func (s staticClientEventsStore) AIAgentClientEvents(
	context.Context,
	AuthorizationResult,
) ([]ClientStreamEvent, error) {
	return s.events, nil
}
