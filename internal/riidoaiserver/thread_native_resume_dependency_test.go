package riidoaiserver

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNativeFollowupProjectionFailureRemainsRetryableAndPrivate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"storage", errors.New("private-upstream-body"), 503}, {"deadline", context.DeadlineExceeded, 504}, {"cancelled", context.Canceled, 408}} {
		t.Run(tc.name, func(t *testing.T) {
			server, source, thread, principal, req := nativeResumeScopeFixture(t)
			source.loadErr = tc.err
			err := server.applyThreadNativeResume(t.Context(), principal, thread, &req)
			if err == nil || req.ResumeSessionID != "" {
				t.Fatal("unavailable authority became a fresh-session approval")
			}
			response := httptest.NewRecorder()
			writeAIAgentClientError(response, err)
			if response.Code != tc.status || strings.Contains(response.Body.String(), "private-upstream-body") {
				t.Fatal("dependency cause lost retryability or disclosed upstream body")
			}
		})
	}
}
