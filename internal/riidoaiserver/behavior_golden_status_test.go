package riidoaiserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func riidoAIServerStatusGoldenPayload(t *testing.T, server http.Handler) string {
	t.Helper()
	var b strings.Builder
	cases := []struct {
		method, path, token string
		want                int
	}{
		{http.MethodGet, "/healthz", "", http.StatusOK},
		{http.MethodGet, "/readyz", "", http.StatusOK},
		{http.MethodGet, "/v1/client/ai-agent/devices", "", http.StatusUnauthorized},
		{http.MethodGet, "/v1/client/ai-agent/devices", "user-token", http.StatusOK},
		{http.MethodGet, "/v1/client/ai-agent/devices/device-dev-macbook/daemons", "user-token", http.StatusOK},
		{http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemons", "user-token", http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		fmt.Fprintf(&b, "%s %s auth=%t status=%d\n", tc.method, tc.path, tc.token != "", resp.Code)
		if resp.Code != tc.want {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, resp.Code, tc.want)
		}
	}
	return b.String()
}
