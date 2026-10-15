package riidoaiserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type observationFailingAuthorizer struct{ err error }

func (a observationFailingAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	return AuthorizationResult{}, a.err
}

func TestRequestFailureObservationBoundsCauseAndCorrelates(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"cancelled", fmt.Errorf("Bearer private: %w", context.Canceled), "cancelled"},
		{"deadline", context.DeadlineExceeded, "deadline_exceeded"},
		{"unknown", errors.New("/Users/private/customer Bearer credential"), "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			defer log.SetOutput(previous)
			req := httptest.NewRequest(http.MethodGet, "/v3/client/workspaces/private-workspace/ai-agent/tasks/private-task/threads?token=secret", nil)
			req = req.WithContext(WithTraceRecorder(req.Context(), &sseCorrelationRecorder{}))
			req.Header.Set("Authorization", "Bearer token")
			_, ok := (Server{config: ServerConfig{Authorizer: observationFailingAuthorizer{tc.err}}}).authorizeRequest(httptest.NewRecorder(), req, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient})
			if ok {
				t.Fatal("unexpected authorization")
			}
			logRequestFailure(req, "projection_reconcile", tc.err)
			got := output.String()
			for _, want := range []string{`stage="authorization"`, `stage="projection_reconcile"`, `error_type="` + tc.want + `"`, `trace_id="12345678901234567890123456789012"`, `route="/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads"`} {
				if !strings.Contains(got, want) {
					t.Fatalf("missing %q: %q", want, got)
				}
			}
			for _, forbidden := range []string{"private-task", "Bearer", "/Users/", "credential", "secret"} {
				if strings.Contains(got, forbidden) {
					t.Fatalf("sensitive value in observation: %q", got)
				}
			}
		})
	}
}
