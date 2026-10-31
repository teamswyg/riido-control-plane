package authpep

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestCLIThreadReadActualHTTPDenialNeverUsesGenericAuthority(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusUnauthorized, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int64
			a, _ := newThreadReadFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				_, _ = io.WriteString(w, threadReadFixtureService+threadReadFixtureToken())
			})
			legacy := &countingAuthorizer{result: AuthorizationResult{PrincipalID: "user-1", WorkspaceID: threadReadFixtureWorkspace}}
			server := httptest.NewServer(riidoaiserver.NewServer(riidoaiserver.ServerConfig{AIAgentClient: newCLIThreadReadIntegrationStore(t), CLIThreadReadAuthorizer: a, Authorizer: legacy}).Handler())
			defer server.Close()
			req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/v3/client/workspaces/"+threadReadFixtureWorkspace+"/ai-agent/tasks/task-fixture/threads", nil)
			req.Header.Set("Authorization", "Bearer "+threadReadFixtureToken())
			response, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(response.Body)
			_ = response.Body.Close()
			want := http.StatusServiceUnavailable
			if status == http.StatusForbidden {
				want = status
			}
			if err != nil || response.StatusCode != want || legacy.calls != 0 || calls.Load() != 1 || response.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("denial retried, downgraded or changed", response.StatusCode, legacy.calls, calls.Load(), err)
			}
			if strings.Contains(string(body), threadReadFixtureService) || strings.Contains(string(body), threadReadFixtureToken()) {
				t.Fatal("upstream credential reflected")
			}
		})
	}
}
