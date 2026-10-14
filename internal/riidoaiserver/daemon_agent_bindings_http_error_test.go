package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPDaemonAgentBindingsRouteErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		server  http.Handler
		request *http.Request
		want    int
	}{
		{
			name:    "wrong method",
			server:  NewServer(ServerConfig{}).Handler(),
			request: httptest.NewRequest(http.MethodPost, "/v1/daemon/agent-bindings", nil),
			want:    http.StatusMethodNotAllowed,
		},
		{
			name:    "unconfigured runtime store",
			server:  NewServer(ServerConfig{}).Handler(),
			request: httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil),
			want:    http.StatusServiceUnavailable,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, tc.request)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestHTTPDaemonAgentBindingsAuthorizationErrors(t *testing.T) {
	runtimeStore := &daemonBindingsRuntimeStore{}
	server := NewServer(ServerConfig{AIAgentClient: runtimeStore}).Handler()
	if resp := serveDaemonBindings(server, "", "secret"); resp.Code != http.StatusUnauthorized {
		t.Fatalf("missing device status=%d body=%s", resp.Code, resp.Body.String())
	}
	if resp := serveDaemonBindings(server, "device-a", ""); resp.Code != http.StatusUnauthorized {
		t.Fatalf("missing secret status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPDaemonAgentBindingsStoreErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{name: "forbidden", err: ErrAuthorizationForbidden, want: http.StatusForbidden},
		{name: "backend", err: errors.New("binding reader failed"), want: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newDaemonBindingsTestServer(&daemonBindingsRuntimeStore{err: tc.err})
			resp := serveDaemonBindings(server, "device-a", "secret")
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
