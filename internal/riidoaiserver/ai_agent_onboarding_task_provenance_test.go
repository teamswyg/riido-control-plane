package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const onboardingCompletionViewerJWT = "viewer-riido-jwt"

// onboardingTaskProvenanceReader answers the monolith provenance question with a fixed result.
type onboardingTaskProvenanceReader struct {
	AIAgentTaskContextReader
	matched bool
	err     error
	bearers []string
}

func (r *onboardingTaskProvenanceReader) IsOnboardingFirstTaskBundleTask(ctx context.Context, bearerToken, workspaceID, taskID string) (bool, error) {
	r.bearers = append(r.bearers, bearerToken)
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return r.matched, r.err
}

func newOnboardingCompletionHTTPTestServer(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	return newOnboardingCompletionHTTPTestServerWith(t, &onboardingTaskProvenanceReader{matched: true})
}

func newOnboardingCompletionHTTPTestServerWith(t *testing.T, provenance *onboardingTaskProvenanceReader) (http.Handler, *Store) {
	t.Helper()
	if provenance.AIAgentTaskContextReader == nil {
		provenance.AIAgentTaskContextReader = &assignmentHTTPRequestTaskContextReader{contextSnapshot: marketingTaskContextFixture()}
	}
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   provenance,
		Authorizer:    authorizer,
	}).Handler()
	return server, assignmentStore
}

func TestHTTPOnboardingCompletionRequiresMonolithConfirmedFirstTaskBundle(t *testing.T) {
	t.Run("monolith says the task is not the viewer's bundle", func(t *testing.T) {
		provenance := &onboardingTaskProvenanceReader{matched: false}
		handler, _ := newOnboardingCompletionHTTPTestServerWith(t, provenance)
		agentID := createOnboardingAgentHTTP(t, handler, "온보딩 에이전트")
		rec := postOnboardingCompletion(t, handler, "task-provenance-1", onboardingCompletionBody(agentID, "schedule"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if len(provenance.bearers) != 1 || provenance.bearers[0] != onboardingCompletionViewerJWT {
			t.Fatalf("monolith must receive the viewer Bearer JWT, got %v", provenance.bearers)
		}
		// Nothing was recorded, so the same agent can still complete a real bundle task.
		provenance.matched = true
		decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-provenance-1", onboardingCompletionBody(agentID, "schedule")))
	})
	t.Run("monolith is unavailable", func(t *testing.T) {
		provenance := &onboardingTaskProvenanceReader{err: errors.New("dashboard metadata returned HTTP 503")}
		handler, _ := newOnboardingCompletionHTTPTestServerWith(t, provenance)
		agentID := createOnboardingAgentHTTP(t, handler, "온보딩 에이전트")
		rec := postOnboardingCompletion(t, handler, "task-provenance-2", onboardingCompletionBody(agentID, "schedule"))
		if rec.Code != http.StatusBadGateway {
			t.Fatalf("status=%d, want 502; body=%s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "503") {
			t.Fatalf("monolith detail leaked into the response: %s", rec.Body.String())
		}
		provenance.err = nil
		provenance.matched = true
		decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-provenance-2", onboardingCompletionBody(agentID, "schedule")))
	})
	t.Run("viewer Bearer JWT missing", func(t *testing.T) {
		provenance := &onboardingTaskProvenanceReader{matched: true}
		handler, _ := newOnboardingCompletionHTTPTestServerWith(t, provenance)
		agentID := createOnboardingAgentHTTP(t, handler, "온보딩 에이전트")
		req := httptest.NewRequest(http.MethodPost, onboardingCompletionWorkspacePath+"task-provenance-3/onboarding-completion", strings.NewReader(onboardingCompletionBody(agentID, "schedule")))
		req.Header.Set(aiAgentTokenHeader, "user-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d, want 401; body=%s", rec.Code, rec.Body.String())
		}
		if len(provenance.bearers) != 0 {
			t.Fatalf("monolith must not be asked without a Bearer JWT: %v", provenance.bearers)
		}
	})
}

func TestRequestBearerTokenIgnoresAgentTokenHeader(t *testing.T) {
	cases := []struct {
		name      string
		headers   map[string]string
		wantToken string
		wantOK    bool
	}{
		{"agent header only", map[string]string{aiAgentTokenHeader: "agent-token"}, "", false},
		{"bearer only", map[string]string{"Authorization": "Bearer viewer-jwt"}, "viewer-jwt", true},
		{"agent header plus bearer", map[string]string{aiAgentTokenHeader: "agent-token", "Authorization": "Bearer viewer-jwt"}, "viewer-jwt", true},
		{"empty bearer", map[string]string{"Authorization": "Bearer   "}, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			token, ok := requestBearerToken(req)
			if token != tc.wantToken || ok != tc.wantOK {
				t.Fatalf("requestBearerToken = %q, %v; want %q, %v", token, ok, tc.wantToken, tc.wantOK)
			}
		})
	}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(aiAgentTokenHeader, "agent-token")
	req.Header.Set("Authorization", "Bearer viewer-jwt")
	if token, _ := requestToken(req); token != "agent-token" {
		t.Fatalf("requestToken must keep preferring the agent header, got %q", token)
	}
}

func TestAIAgentPrivateTaskContextOnboardingBundleCheckFailsClosed(t *testing.T) {
	cases := []struct {
		name        string
		mode        string
		bearer      string
		want        bool
		wantErr     bool
		wantQueries int
	}{
		{name: "own bundle task", mode: "true", want: true, wantQueries: 2},
		{name: "monolith says no", mode: "false", want: false, wantQueries: 2},
		{name: "monolith HTTP error", mode: "monolith_error", wantErr: true, wantQueries: 1},
		{name: "task not visible to viewer", mode: "forbidden_location", want: false, wantQueries: 1},
		{name: "task in another workspace", mode: "other_workspace", want: false, wantQueries: 1},
		{name: "answer field missing", mode: "missing_answer", wantErr: true, wantQueries: 2},
		{name: "empty bearer makes no request", mode: "true", bearer: " ", want: false, wantQueries: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				queries++
				if r.Method != http.MethodPost || r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer "+onboardingCompletionViewerJWT {
					t.Errorf("unexpected monolith request: %s %s auth=%q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
				}
				var body struct {
					Query     string            `json:"query"`
					Variables map[string]string `json:"variables"`
				}
				_ = json.NewDecoder(r.Body).Decode(&body)
				if strings.Contains(body.Query, "OnboardingTaskLocation") {
					switch tc.mode {
					case "monolith_error":
						w.WriteHeader(http.StatusInternalServerError)
						return
					case "forbidden_location":
						_, _ = w.Write([]byte(`{"errors":[{"extensions":{"code":"FORBIDDEN"}}]}`))
						return
					}
					workspace := "workspace-a"
					if tc.mode == "other_workspace" {
						workspace = "workspace-b"
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentWorkspace": map[string]any{
						"id": "task-1", "componentType": "task", "team": map[string]any{"id": "team-1", "workspace": map[string]string{"id": workspace}},
					}}})
					return
				}
				if !strings.Contains(body.Query, "isOnboardingFirstTaskBundleTask") || body.Variables["teamId"] != "team-1" || body.Variables["taskId"] != "task-1" {
					t.Errorf("provenance query changed: %+v", body)
				}
				switch tc.mode {
				case "missing_answer":
					_, _ = w.Write([]byte(`{"data":{}}`))
				default:
					answer := tc.mode == "true"
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"isOnboardingFirstTaskBundleTask": answer}})
				}
			}))
			defer server.Close()
			client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			bearer := onboardingCompletionViewerJWT
			if tc.bearer != "" {
				bearer = tc.bearer
			}
			got, err := client.IsOnboardingFirstTaskBundleTask(context.Background(), bearer, "workspace-a", "task-1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("answer = %v, want %v", got, tc.want)
			}
			if queries != tc.wantQueries {
				t.Fatalf("monolith queries = %d, want %d", queries, tc.wantQueries)
			}
		})
	}
}
