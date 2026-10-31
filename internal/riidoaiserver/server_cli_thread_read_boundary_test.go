package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIThreadReadCredentialNeverFallsBack(t *testing.T) {
	for _, action := range []AuthorizationAction{AuthorizationActionRead, AuthorizationActionCreate, AuthorizationActionStream} {
		t.Run(string(action), func(t *testing.T) {
			calls := 0
			s := NewServer(ServerConfig{Authorizer: authorizationFunc(func(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
				calls++
				return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: req.WorkspaceID}, nil
			})})
			original := "/v3/client/workspaces/workspace-fixture/ai-agent/tasks/task-fixture/threads"
			r := httptest.NewRequest(http.MethodGet, original, nil)
			r.Header.Set("Authorization", "Bearer riido_cli_thread_v1_"+strings.Repeat("A", 43))
			r = requestWithAIAgentWorkspaceIDAndPath(r, "workspace-fixture", "/v3/client/ai-agent/tasks/task-fixture/threads")
			w := httptest.NewRecorder()
			_, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: action, WorkspaceID: "workspace-fixture", TaskID: "task-fixture"})
			if ok || calls != 0 || w.Code < 400 {
				t.Fatalf("CLI ticket reached general authorizer: allowed=%v calls=%d status=%d", ok, calls, w.Code)
			}
		})
	}
}
