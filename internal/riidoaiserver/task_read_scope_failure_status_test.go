package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// RIID-6008: read-scope failures that retrying cannot fix are reported with a
// non-retryable status; transient and other task-context failures keep 503.
func TestTaskReadScopeFailureStatus(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want int
	}{
		"denied":           {latestTaskDocumentError{taskReadScopeError{taskContextReadError{"denied"}}}, http.StatusForbidden},
		"missing":          {latestTaskDocumentError{taskReadScopeError{taskContextReadError{"missing"}}}, http.StatusNotFound},
		"integrity":        {latestTaskDocumentError{taskReadScopeError{taskContextReadError{"integrity"}}}, http.StatusUnprocessableEntity},
		"unavailable":      {latestTaskDocumentError{taskReadScopeError{taskContextReadError{"unavailable"}}}, http.StatusServiceUnavailable},
		"timeout":          {latestTaskDocumentError{taskReadScopeError{context.DeadlineExceeded}}, http.StatusServiceUnavailable},
		"provider denied":  {latestTaskDocumentError{taskContextReadError{"denied"}}, http.StatusServiceUnavailable},
		"provider timeout": {latestTaskDocumentError{context.DeadlineExceeded}, http.StatusServiceUnavailable},
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			writeAIAgentClientError(response, tc.err)
			if response.Code != tc.want {
				t.Fatalf("status=%d want %d", response.Code, tc.want)
			}
		})
	}
}

func TestTaskReadScopeGraphQLClassifiesIntegrity(t *testing.T) {
	for name, tc := range map[string]struct {
		extensions map[string]string
		want       string
	}{
		"integrity":        {map[string]string{"code": "conflict", "reason": readScopeIntegrityReason}, "integrity"},
		"version conflict": {map[string]string{"code": "conflict"}, "unavailable"},
		"other reason":     {map[string]string{"code": "conflict", "reason": "something_else"}, "unavailable"},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": nil}, "errors": []any{map[string]any{"message": "private upstream detail", "extensions": tc.extensions}}})
			}))
			t.Cleanup(server.Close)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true})
			var out taskReadScope
			err := client.readScopeGraphQL(t.Context(), "synthetic", readScopeQuery, "getComponentReadScope", map[string]any{"componentId": "root", "teamId": "team"}, &out)
			var failure taskContextReadError
			if !errors.As(err, &failure) || failure.outcome != tc.want {
				t.Fatalf("got %v want %s", err, tc.want)
			}
		})
	}
}
