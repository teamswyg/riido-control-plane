package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAssignmentReadScopeRetainedViewerUsesPrivateTypedAPIWithCaptureOff(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer current-viewer" {
			t.Error("current viewer token missing")
		}
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace-dev-riid"}}}`))
			return
		}
		if r.URL.Path != "/graphql" {
			http.Error(w, "legacy context forbidden in this fixture", http.StatusForbidden)
			return
		}
		calls.Add(1)
		scope := readScopeCaptureInitial()
		scope["works"].([]any)[0].(map[string]any)["body"] = readScopeCaptureBody(0, "helloworld", nil, true)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": scope}})
	}))
	t.Cleanup(upstream.Close)
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	if err != nil {
		t.Fatal(err)
	}
	request := AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: defaultAIAgentClientWorkspaceID, BearerToken: "current-viewer", capturedReadScope: true}
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), request)
	if err != nil || got.ReadScope == nil || calls.Load() != 1 || client.assignmentReadScope {
		t.Fatal("stored-scope selection did not use current typed API independently of capture flag", err)
	}
	request.capturedReadScope = false
	_, _ = client.GetAIAgentTaskContextForRequest(t.Context(), request)
	if calls.Load() != 1 {
		t.Fatal("ordinary flag-off capture was promoted to typed scope")
	}
}
