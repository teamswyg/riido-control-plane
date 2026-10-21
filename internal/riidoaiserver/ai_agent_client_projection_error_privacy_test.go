package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLegacyClientProjectionErrorsArePrivateAndTyped(t *testing.T) {
	for _, path := range []string{
		"/v1/client/ai-agent/bootstrap",
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/bootstrap",
		taskThreadsErrorTestPath(), subscriptionErrorTestPath(),
	} {
		for _, tc := range []struct {
			name   string
			err    error
			status int
		}{
			{"cancelled", fmt.Errorf("PRIVATE_PROJECTION_CANARY: %w", context.Canceled), 408},
			{"deadline", fmt.Errorf("PRIVATE_PROJECTION_CANARY: %w", context.DeadlineExceeded), 504},
			{"dependency", errors.New("PRIVATE_PROJECTION_CANARY upstream body"), 502},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				assignment := NewStore()
				t.Cleanup(assignment.Close)
				store := bootstrapDevicesErrorStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), reconcileErr: tc.err}
				server := newTaskThreadReadErrorTestServer(t, []string{"ai-agent:*"}, store, assignment)
				req := httptest.NewRequest(http.MethodGet, path, nil)
				req.Header.Set("Authorization", "Bearer user-token")
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				if resp.Code != tc.status || strings.Contains(resp.Body.String(), "PRIVATE_PROJECTION_CANARY") {
					t.Fatalf("projection failure was misclassified or leaked raw detail: status=%d want=%d", resp.Code, tc.status)
				}
			})
		}
	}
}
