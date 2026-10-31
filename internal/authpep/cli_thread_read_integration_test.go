package authpep

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestCLIThreadReadActualHTTPPreservesVisibilityAndIdentifiers(t *testing.T) {
	for _, tc := range []struct {
		name, user string
		roles      []string
		count      int
	}{
		{"owner", "user-1", []string{}, 2}, {"other owner", "user-2", []string{}, 2}, {"viewer", "user-3", []string{}, 1}, {"admin", "user-3", []string{"admin"}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var checks atomic.Int64
			a, _ := newThreadReadFixture(t, func(w http.ResponseWriter, r *http.Request) {
				assertThreadReadFixtureRequest(t, r)
				checks.Add(1)
				reply := threadReadFixtureResponse()
				reply.PrincipalID = tc.user
				reply.Roles = tc.roles
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(reply)
			})
			store := newCLIThreadReadIntegrationStore(t)
			legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "wrong"}}
			server := httptest.NewServer(riidoaiserver.NewServer(riidoaiserver.ServerConfig{AIAgentClient: store, CLIThreadReadAuthorizer: a, Authorizer: legacy}).Handler())
			defer server.Close()
			for range 2 {
				req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/v3/client/workspaces/"+threadReadFixtureWorkspace+"/ai-agent/tasks/task-fixture/threads", nil)
				req.Header.Set("Authorization", "Bearer "+threadReadFixtureToken())
				response, err := server.Client().Do(req)
				if err != nil {
					t.Fatal("CP HTTP failed", err)
				}
				body, err := io.ReadAll(response.Body)
				_ = response.Body.Close()
				var history riidoaiserver.AIAgentTaskThreadHistoryCollectionResponse
				if err != nil || response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "no-store" || json.Unmarshal(body, &history) != nil || len(history.Threads) != tc.count {
					t.Fatal("history visibility changed", response.StatusCode, len(history.Threads))
				}
				for _, thread := range history.Threads {
					if thread.ThreadID == "" || thread.ConversationID == "" || thread.AgentSnapshotID == "" || thread.AssignmentID == "" || history.AgentSnapshots[thread.AgentSnapshotID] == nil {
						t.Fatal("history identifiers collapsed")
					}
				}
			}
			if checks.Load() != 2 || legacy.calls != 0 {
				t.Fatal("current check cached or generic fallback used", checks.Load(), legacy.calls)
			}
		})
	}
}

func newCLIThreadReadIntegrationStore(t *testing.T) *riidoaiserver.DevelopmentAIAgentClientStore {
	t.Helper()
	store := riidoaiserver.NewDevelopmentAIAgentClientStore()
	for _, pair := range []struct{ user, agent string }{{"user-1", "agent-owned-codex"}, {"user-2", "agent-private-cursor"}, {"user-2", "agent-public-openclaw"}} {
		_, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), AuthorizationResult{PrincipalID: pair.user, WorkspaceID: threadReadFixtureWorkspace}, "task-fixture", riidoaiserver.AssignAIAgentTaskRequest{AgentID: pair.agent})
		if err != nil {
			t.Fatal("thread fixture failed", err)
		}
	}
	return store
}
