package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlPlaneOwnerReceiverIsSourceReadyButNotRegistered(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/owner/graphql", strings.NewReader(
		`{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewServer(ServerConfig{}).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("source-ready receiver must stay unregistered, status=%d", response.Code)
	}
}
