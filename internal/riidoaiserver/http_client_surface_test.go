package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceHTTPClientSurfaceClassifiesKnownSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		route     string
		userAgent string
		want      string
	}{
		{name: "daemon poll", route: "/v1/agents/{agent_id}/poll", want: "daemon"},
		{name: "daemon snapshot", route: "/v1/daemon/runtime-snapshot", want: "daemon"},
		{name: "client v3", route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads", want: "client_app"},
		{name: "desktop explicit user agent", route: "/v2/client/workspaces/{workspace_id}/ai-agent/bootstrap", userAgent: "Riido-Desktop/0.0.15", want: "desktop"},
		{name: "desktop electron shell", route: "/v2/client/workspaces/{workspace_id}/ai-agent/bootstrap", userAgent: "Mozilla/5.0 Electron/37.2.0 Chrome/138.0.0.0", want: "desktop"},
		{name: "desktop candidate devices", route: "/v2/client/workspaces/{workspace_id}/ai-agent/devices", want: "desktop_candidate"},
		{name: "component", route: "/v1/component-tasks/{task_id}", want: "component_integration"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := traceHTTPClientSurface(tt.route, tt.route, tt.userAgent); got != tt.want {
				t.Fatalf("surface = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTTPTracingRecordsClientSurface(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	handler := withHTTPTracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), recorder)
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	spans := recorder.snapshot()
	if len(spans) != 1 {
		t.Fatalf("spans = %+v", spans)
	}
	if got := spans[0].Attributes[riidoClientSurfaceTraceKey]; got != "daemon" {
		t.Fatalf("client surface = %q, attrs=%+v", got, spans[0].Attributes)
	}
}
