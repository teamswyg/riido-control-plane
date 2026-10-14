package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientBootstrapAndDevicesErrors(t *testing.T) {
	errorStore := func(bootstrapErr, devicesErr error) AIAgentClientStore {
		return bootstrapDevicesErrorStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), bootstrapErr: bootstrapErr, devicesErr: devicesErr}
	}
	for _, tc := range []struct {
		name, method, path, wantBody string
		server                       http.Handler
		want                         int
	}{
		{"bootstrap missing store", http.MethodGet, "/v1/client/ai-agent/bootstrap", "ai agent client store is not configured", NewServer(ServerConfig{}).Handler(), http.StatusServiceUnavailable},
		{"bootstrap wrong method", http.MethodPost, "/v1/client/ai-agent/bootstrap", "method not allowed", bootstrapDevicesErrorServer(t, errorStore(nil, nil), nil), http.StatusMethodNotAllowed},
		{"bootstrap backend error", http.MethodGet, "/v1/client/ai-agent/bootstrap", "bootstrap failed", bootstrapDevicesErrorServer(t, errorStore(errors.New("bootstrap failed"), nil), nil), http.StatusBadRequest},
		{"devices missing store", http.MethodGet, "/v1/client/ai-agent/devices", "ai agent client store is not configured", NewServer(ServerConfig{}).Handler(), http.StatusServiceUnavailable},
		{"devices wrong method", http.MethodPost, "/v1/client/ai-agent/devices", "method not allowed", bootstrapDevicesErrorServer(t, errorStore(nil, nil), nil), http.StatusMethodNotAllowed},
		{"devices backend error", http.MethodGet, "/v1/client/ai-agent/devices", "devices failed", bootstrapDevicesErrorServer(t, errorStore(nil, errors.New("devices failed")), nil), http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer user-token")
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, req)
			if resp.Code != tc.want || !strings.Contains(resp.Body.String(), tc.wantBody) {
				t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
			}
		})
	}
}

func TestHTTPAIAgentClientBootstrapReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)
	store := bootstrapDevicesErrorStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), reconcileErr: errors.New("projection reconcile failed")}
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	bootstrapDevicesErrorServer(t, store, assignment).ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway || !strings.Contains(resp.Body.String(), "projection reconcile failed") {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}
