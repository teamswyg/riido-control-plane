package authpep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestCLIAgentReadActualHTTPVisibilityAndProjection(t *testing.T) {
	for _, tc := range []struct {
		name, user string
		roles      []string
		count      int
	}{{"owner", "user-1", []string{}, 3}, {"other owner", "user-2", []string{}, 2}, {"viewer", "user-3", []string{}, 1}, {"admin", "user-3", []string{"admin"}, 4}} {
		t.Run(tc.name, func(t *testing.T) {
			var checks atomic.Int64
			a, _ := newAgentReadFixture(t, func(w http.ResponseWriter, r *http.Request) {
				assertAgentReadFixtureRequest(t, r)
				checks.Add(1)
				reply := agentReadFixtureResponse()
				reply.PrincipalID, reply.Roles = tc.user, tc.roles
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(reply)
			})
			legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "wrong"}}
			server := httptest.NewServer(riidoaiserver.NewServer(riidoaiserver.ServerConfig{AIAgentClient: riidoaiserver.NewDevelopmentAIAgentClientStore(), CLIAgentReadAuthorizer: a, Authorizer: legacy}).Handler())
			defer server.Close()
			for _, suffix := range []string{"", "?cursor=1", "/agent-public-openclaw"} {
				raw := agentDirectoryHTTP(t, server, suffix)
				for _, secret := range []string{"instruction", "owner_principal_id", "user-1", "user-2", "device_secret", "lease_token"} {
					if strings.Contains(string(raw), secret) {
						t.Fatal("nonpublic field leaked", secret)
					}
				}
				if suffix == "/agent-public-openclaw" {
					var detail riidoaiserver.AgentDirectoryDetail
					if json.Unmarshal(raw, &detail) != nil || detail.Agent.AgentID != "agent-public-openclaw" || detail.Agent.IsCurrentAgent || detail.PolicyChecked || detail.FetchedAt.IsZero() {
						t.Fatal("invalid public detail")
					}
					continue
				}
				var page riidoaiserver.AgentDirectoryPage
				if json.Unmarshal(raw, &page) != nil || page.Total != tc.count || page.PolicyChecked || page.FetchedAt.IsZero() || (suffix == "" && len(page.Agents) != tc.count) || (suffix != "" && len(page.Agents) != 0) {
					t.Fatal("visibility or pagination changed")
				}
				for _, entry := range page.Agents {
					if entry.IsCurrentAgent || entry.Callable != (len(entry.UnavailableReasons) == 0) || entry.Name == "" {
						t.Fatal("user lookup claimed a run or lost metadata")
					}
				}
			}
			if checks.Load() != 3 || legacy.calls != 0 {
				t.Fatal("cached authority or generic fallback", checks.Load(), legacy.calls)
			}
		})
	}
}
