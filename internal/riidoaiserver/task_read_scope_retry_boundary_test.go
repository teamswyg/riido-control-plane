package riidoaiserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeDoesNotRetryGraphQLOrMalformedResponses(t *testing.T) {
	for _, body := range []string{
		`{"errors":[{"extensions":{"code":"FORBIDDEN"}}]}`,
		`{"errors":[{"extensions":{"code":"conflict","reason":"read_scope_integrity"}}]}`,
		`{"errors":[{"extensions":{"code":"conflict"}}]}`,
		`{"errors":[{"extensions":{"code":"internal"}}]}`,
		`{"data":{"getComponentReadScope":null}}`,
		`{`,
	} {
		t.Run(body, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				_, _ = io.WriteString(w, body)
			}))
			t.Cleanup(server.Close)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
			var out taskReadScope
			err := client.readScopeGraphQL(t.Context(), "synthetic", readScopeQuery, "getComponentReadScope", nil, &out)
			if err == nil || calls.Load() != 1 {
				t.Fatalf("nontransient response retried or accepted: calls=%d err=%v", calls.Load(), err)
			}
		})
	}
}
