package riidoaiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskReadScopeGraphQLClassifiesProducerErrorCodes(t *testing.T) {
	for code, want := range map[string]string{"authentication_required": "denied", "permission_denied": "denied", "not_found": "missing", "UNAUTHENTICATED": "denied", "FORBIDDEN": "denied", "service_unavailable": "unavailable"} {
		t.Run(code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": nil}, "errors": []any{map[string]any{"message": "private upstream detail", "extensions": map[string]string{"code": code}}}})
			}))
			t.Cleanup(server.Close)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true})
			var out taskReadScope
			err := client.readScopeGraphQL(t.Context(), "synthetic", readScopeQuery, "getComponentReadScope", map[string]any{"componentId": "root", "teamId": "team"}, &out)
			var failure taskContextReadError
			if !errors.As(err, &failure) || failure.outcome != want {
				t.Fatalf("got %v want %s", err, want)
			}
		})
	}
}
