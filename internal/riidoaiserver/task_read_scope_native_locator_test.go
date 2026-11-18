package riidoaiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskReadScopeNativeInitialLocatorFailsClosed(t *testing.T) {
	for _, name := range []string{"denied", "not found", "missing workspace", "missing bearer", "missing task", "scope workspace", "scope root", "missing team", "missing root team", "old root team", "downgrade"} {
		t.Run(name, func(t *testing.T) {
			capture := nativeTestCapture(t, "component")
			req := AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace", BearerToken: "human"}
			mutateNativeLocatorCase(name, capture, &req)
			calls := 0
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/graphql" {
					t.Error("native locator attempted legacy REST")
				}
				switch name {
				case "denied":
					_, _ = w.Write([]byte(`{"errors":[{"message":"denied","extensions":{"code":"permission_denied"}}]}`))
				case "not found":
					http.NotFound(w, r)
				default:
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": capture.Scope}})
				}
			}))
			defer upstream.Close()
			client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
			if err != nil {
				t.Fatal(err)
			}
			server := NewServer(ServerConfig{TaskContext: client})
			result, err := server.assignRequestWithTaskContextPromptForClientResult(t.Context(), req.ComponentID, AssignRequest{Prompt: "Inspect only"}, req)
			var unavailable latestTaskDocumentError
			if !errors.As(err, &unavailable) || result.TaskContext != nil || result.Request.Prompt != "" {
				t.Fatalf("native lookup produced context-free execution: result=%+v err=%v", result, err)
			}
			want := 1
			if name == "missing workspace" || name == "missing bearer" || name == "missing task" {
				want = 0
			}
			if calls != want {
				t.Fatalf("unexpected retries: calls=%d want=%d", calls, want)
			}
		})
	}
}
