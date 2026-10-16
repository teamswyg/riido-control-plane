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

func TestV3ThreadReconcileSeparatesCancellationDeadlineAndDependency(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"cancelled", context.Canceled, 408},
		{"wrapped cancellation", fmt.Errorf("private-marker: %w", context.Canceled), 408},
		{"deadline", fmt.Errorf("private-marker: %w", context.DeadlineExceeded), 504},
		{"dependency", errors.New("private-marker upstream body"), 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assignment := NewStore()
			t.Cleanup(assignment.Close)
			server := newTaskThreadReadErrorTestServer(t, []string{"ai-agent:*"}, v3HistoryErrorStore{reconcileErr: tc.err}, assignment)
			req := httptest.NewRequest(http.MethodGet, v3HistoryErrorTestPath(), nil)
			req.Header.Set("Authorization", "Bearer user-token")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.status {
				t.Fatalf("reconcile status=%d want=%d", resp.Code, tc.status)
			}
			if strings.Contains(resp.Body.String(), "private-marker") || strings.Contains(resp.Body.String(), "upstream body") {
				t.Fatal("upstream reconcile error body crossed the HTTP boundary")
			}
		})
	}
}
