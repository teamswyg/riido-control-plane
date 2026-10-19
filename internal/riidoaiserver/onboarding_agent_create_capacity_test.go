package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOnboardingAgentCreateReceiptCapacityFailsClosedWithoutEviction(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	now := time.Now().UTC()
	store.onboardingAgentCreates = map[string]onboardingAgentCreateReceipt{}
	for i := 0; ; i++ {
		key := fmt.Sprintf("%064x", i)
		store.onboardingAgentCreates[key] = onboardingAgentCreateReceipt{Fingerprint: strings.Repeat("a", 64), AgentID: "agent-owned-claude", CreatedAt: now}
		raw, err := json.Marshal(store.onboardingAgentCreates)
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) > onboardingAgentCreateReceiptBytes {
			delete(store.onboardingAgentCreates, key)
			break
		}
	}
	beforeReceipts, beforeAgents := len(store.onboardingAgentCreates), len(store.agents)
	_, err := store.CreateAIAgentFromOnboardingFixture(withOnboardingAgentCreateRequestID(t.Context(), "create-at-capacity"), profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("long result "+strings.Repeat("z", 200)))
	if err == nil {
		t.Fatal("receipt capacity was exceeded without refusing creation")
	}
	rec := httptest.NewRecorder()
	writeAIAgentClientError(rec, err)
	if rec.Code != http.StatusServiceUnavailable || len(store.onboardingAgentCreates) != beforeReceipts || len(store.agents) != beforeAgents {
		t.Fatalf("capacity error status=%d receipt delta=%d agent delta=%d, want503 and no eviction/creation", rec.Code, len(store.onboardingAgentCreates)-beforeReceipts, len(store.agents)-beforeAgents)
	}
}
