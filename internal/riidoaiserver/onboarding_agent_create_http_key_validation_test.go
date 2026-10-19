package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestOnboardingAgentCreateHTTPRejectsKeyConflictAndInvalidKeys(t *testing.T) {
	for _, changed := range []string{"payload", "fixture"} {
		t.Run(changed, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			handler := onboardingAgentCreateHTTPHandler(t, store)
			body := onboardingAgentCreateHTTPBody(t, "fixed-create-key")
			decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token"))
			before := len(store.agents)
			path := onboardingAgentCreatePath
			if changed == "payload" {
				body = strings.Replace(body, "response-loss fixture", "different draft", 1)
			} else {
				path = strings.Replace(path, "riido_pm", "yeongsil_backend", 1)
			}
			if rec := postOnboardingAgentCreateHTTP(handler, path, body, "ai-agent-token"); rec.Code != http.StatusConflict {
				t.Fatalf("same key with changed %s status=%d, want 409", changed, rec.Code)
			}
			if len(store.agents) != before {
				t.Fatal("key conflict created another agent")
			}
		})
	}
	for _, key := range []string{" ", "bad/key", "with space", "한글", strings.Repeat("x", 129)} {
		t.Run(key, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			handler := onboardingAgentCreateHTTPHandler(t, store)
			before := len(store.agents)
			rec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, onboardingAgentCreateHTTPBody(t, key), "ai-agent-token")
			if rec.Code != http.StatusBadRequest || len(store.agents) != before {
				t.Fatalf("invalid key status=%d delta=%d, want 400 without creation", rec.Code, len(store.agents)-before)
			}
		})
	}
}

func TestOnboardingAgentCreateHTTPRejectsExplicitEmptyKey(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	handler := onboardingAgentCreateHTTPHandler(t, store)
	body := strings.Replace(onboardingAgentCreateHTTPBody(t, "nonempty"), `"request_id":"nonempty"`, `"request_id":""`, 1)
	before := len(store.agents)
	rec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token")
	if rec.Code != http.StatusBadRequest || len(store.agents) != before {
		t.Fatalf("explicit empty request key status=%d delta=%d, want400 withoutcreation", rec.Code, len(store.agents)-before)
	}
}

func TestOnboardingAgentCreateHTTPRejectsNullKey(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	handler := onboardingAgentCreateHTTPHandler(t, store)
	body := strings.Replace(onboardingAgentCreateHTTPBody(t, "nonempty"), `"request_id":"nonempty"`, `"request_id":null`, 1)
	before := len(store.agents)
	rec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token")
	if rec.Code != http.StatusBadRequest || len(store.agents) != before {
		t.Fatalf("null request key status=%d delta=%d, want400 withoutcreation", rec.Code, len(store.agents)-before)
	}
}
