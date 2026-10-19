package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestOnboardingAgentCreatePersistentRejectedCommitDoesNotPublishPhantom(t *testing.T) {
	snapshots := &onboardingAgentCreateFaultSnapshot{}
	store := onboardingAgentCreatePersistent(t, snapshots)
	before := len(store.agents)
	snapshots.failBefore = true
	handler := onboardingAgentCreateHTTPHandler(t, store)
	body := onboardingAgentCreateHTTPBody(t, "rejected-before-commit")
	rec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token")
	if rec.Code != http.StatusServiceUnavailable || len(store.agents) != before {
		t.Fatalf("rejected commit status=%d local delta=%d, want 503 without a phantom agent", rec.Code, len(store.agents)-before)
	}
	snapshots.failBefore = false
	restarted := onboardingAgentCreatePersistent(t, snapshots)
	if len(restarted.agents) != before {
		t.Fatal("rejected create was present in durable state")
	}
	first := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(onboardingAgentCreateHTTPHandler(t, restarted), onboardingAgentCreatePath, body, "ai-agent-token"))
	replayed := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(onboardingAgentCreateHTTPHandler(t, store), onboardingAgentCreatePath, body, "ai-agent-token"))
	if first.Agent.AgentID != replayed.Agent.AgentID || len(store.agents) != before+1 {
		t.Fatal("retry after rejected commit failed to converge to the committed creation")
	}
}
