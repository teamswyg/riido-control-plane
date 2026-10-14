package controlplane

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestControlPlaneOwnerReceiverFailsClosedWithoutExplicitHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/owner/graphql", strings.NewReader(
		`{"query":"query ControlPlaneOwnerHealthCheck { healthCheck }","operationName":"ControlPlaneOwnerHealthCheck"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	riidoaiserver.NewServer(riidoaiserver.ServerConfig{}).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing receiver must fail closed, status=%d", response.Code)
	}
}
