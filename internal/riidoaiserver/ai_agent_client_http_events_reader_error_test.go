package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientEventsReaderOnlyError(t *testing.T) {
	server := newClientEventsErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		clientEventsReaderOnlyErrorStore{
			AIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			err:                errors.New("event reader failed"),
		},
	)
	req := httptest.NewRequest(http.MethodGet, clientEventsErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("events status=%d want=%d body=%s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}
