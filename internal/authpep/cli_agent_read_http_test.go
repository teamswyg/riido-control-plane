package authpep

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestCLIAgentReadPurposeAndCurrentAuthority(t *testing.T) {
	var calls atomic.Int64
	a, _ := newAgentReadFixture(t, func(w http.ResponseWriter, r *http.Request) {
		assertAgentReadFixtureRequest(t, r)
		if calls.Add(1) == 3 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agentReadFixtureResponse())
	})
	for range 2 {
		if result, err := a.Authorize(t.Context(), agentReadFixtureToken(), agentReadFixtureRequest()); err != nil || result.PrincipalID != "user-1" {
			t.Fatal("valid current agent authority rejected", err)
		}
	}
	if _, err := a.Authorize(t.Context(), agentReadFixtureToken(), agentReadFixtureRequest()); !errors.Is(err, ErrAuthorizationForbidden) {
		t.Fatal("revoked grant was accepted", err)
	}
	for _, tc := range []struct {
		token string
		req   AuthorizationRequest
	}{{threadReadFixtureToken(), agentReadFixtureRequest()}, {agentReadFixtureToken(), threadReadFixtureRequest()}} {
		if _, err := a.Authorize(t.Context(), tc.token, tc.req); !errors.Is(err, ErrAuthorizationForbidden) {
			t.Fatal("cross-purpose credential reached service", err)
		}
	}
	if calls.Load() != 3 {
		t.Fatal("unexpected service calls", calls.Load())
	}
	thread, _ := newThreadReadFixture(t, func(http.ResponseWriter, *http.Request) { t.Error("agent ticket reached thread service") })
	if _, err := thread.Authorize(t.Context(), agentReadFixtureToken(), threadReadFixtureRequest()); !errors.Is(err, ErrAuthorizationForbidden) {
		t.Fatal("agent ticket crossed into threads", err)
	}
	wrong, _ := json.Marshal(threadReadFixtureResponse())
	if _, err := a.decode(wrong, agentReadFixtureRequest()); err == nil {
		t.Fatal("thread decision authorized agent directory")
	}
	wrong, _ = json.Marshal(agentReadFixtureResponse())
	if _, err := thread.decode(wrong, threadReadFixtureRequest()); err == nil {
		t.Fatal("agent decision authorized threads")
	}
}
