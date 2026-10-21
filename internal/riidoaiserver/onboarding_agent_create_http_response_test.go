package riidoaiserver

import (
	"net/http"
	"reflect"
	"testing"
)

func TestOnboardingAgentCreateHTTPResponseLossReplaysOneAgent(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	handler := onboardingAgentCreateHTTPHandler(t, store)
	before := len(store.agents)
	body := onboardingAgentCreateHTTPBody(t, "onboarding-create-response-loss")
	// The server commits the first request, but the caller loses its response.
	first := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token"))
	replayed := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token"))
	if !reflect.DeepEqual(first, replayed) || len(store.agents) != before+1 {
		t.Fatalf("response-loss retry created a second agent: first=%q replay=%q count delta=%d", first.Agent.AgentID, replayed.Agent.AgentID, len(store.agents)-before)
	}
	if rec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("replay bypassed authorization: status=%d", rec.Code)
	}
}

func TestOnboardingAgentCreateHTTPPreservesUnkeyedAndNewKeyCreates(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	handler := onboardingAgentCreateHTTPHandler(t, store)
	seen := map[string]bool{}
	for _, key := range []string{"", "", "new-create-key-1", "new-create-key-2"} {
		response := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, onboardingAgentCreateHTTPBody(t, key), "ai-agent-token"))
		if seen[response.Agent.AgentID] {
			t.Fatal("intentional separate create with same name was deduplicated")
		}
		seen[response.Agent.AgentID] = true
	}
}

func TestOnboardingAgentCreateHTTPV1AndV2ReplaySameOperation(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	handler := onboardingAgentCreateHTTPHandler(t, store)
	body := onboardingAgentCreateHTTPBody(t, "cross-alias-create")
	first := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, "/v1/client/ai-agent/onboarding/fixtures/riido_pm/agents", body, "ai-agent-token"))
	replay := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token"))
	if replay.Agent.AgentID != first.Agent.AgentID {
		t.Fatal("v1/v2 aliases created separate agents for one request key")
	}
}
