package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleAgentPollUsesLongPollWithCappedHold(t *testing.T) {
	var gotHold, gotTick time.Duration
	store := &handlerLongPollStore{
		handlerAssignmentStore: &handlerAssignmentStore{},
		wait: func(_ context.Context, agentID string, req PollRequest, hold, tick time.Duration) (PollResponse, error) {
			if agentID != "agent-a" || req.WaitMs != 1000 {
				t.Fatalf("poll target = agent:%s wait:%d", agentID, req.WaitMs)
			}
			gotHold, gotTick = hold, tick
			return PollResponse{SchemaVersion: SchemaVersion, Action: PollNone}, nil
		},
	}
	server := Server{
		assignment: store,
		config: ServerConfig{
			Authorizer:      assignmentHTTPAuthorizer(t, []string{"agent:agent-a:poll"}),
			LongPollMaxHold: 20 * time.Millisecond,
			LongPollTick:    3 * time.Millisecond,
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"wait_ms":1000}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentPoll(resp, req, "agent-a")
	if resp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", resp.Code, resp.Body.String())
	}
	if gotHold != 20*time.Millisecond || gotTick != 3*time.Millisecond {
		t.Fatalf("hold=%s tick=%s", gotHold, gotTick)
	}
}

func TestHandleAgentPollStoreErrorFailsClosed(t *testing.T) {
	store := &handlerAssignmentStore{
		poll: func(context.Context, string, PollRequest) (PollResponse, error) {
			return PollResponse{}, errors.New("binding validation failed")
		},
	}
	server := Server{
		assignment: store,
		config:     ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:poll"})},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.handleAgentPoll(resp, req, "agent-a")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("poll error status=%d body=%s", resp.Code, resp.Body.String())
	}
}
