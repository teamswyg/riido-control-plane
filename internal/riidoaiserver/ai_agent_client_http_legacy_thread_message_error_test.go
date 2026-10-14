package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientLegacyThreadMessageRouteErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		server http.Handler
		method string
		path   string
		want   int
	}{
		{
			name:   "unconfigured store",
			server: NewServer(ServerConfig{}).Handler(),
			method: http.MethodPost,
			path:   "/v1/client/ai-agent/threads/thread-a/messages",
			want:   http.StatusServiceUnavailable,
		},
		{
			name:   "malformed route",
			server: newTaskThreadReadErrorTestServer(t, nil, NewDevelopmentAIAgentClientStore(), nil),
			method: http.MethodPost,
			path:   "/v1/client/ai-agent/threads/",
			want:   http.StatusNotFound,
		},
		{
			name:   "wrong method",
			server: newTaskThreadReadErrorTestServer(t, nil, NewDevelopmentAIAgentClientStore(), nil),
			method: http.MethodGet,
			path:   "/v1/client/ai-agent/threads/thread-a/messages",
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "unknown suffix",
			server: newTaskThreadReadErrorTestServer(t, nil, NewDevelopmentAIAgentClientStore(), nil),
			method: http.MethodPost,
			path:   "/v1/client/ai-agent/threads/thread-a/events",
			want:   http.StatusNotFound,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
