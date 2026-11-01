package authpep

import (
	"net/http"
	"strings"
	"testing"
)

func TestCLIAgentReadHTTPRejectsDependencyFailures(t *testing.T) {
	for _, tc := range []struct {
		name          string
		status        int
		content, body string
	}{
		{"unauthorized service", 401, "application/json", `{}`},
		{"unavailable", 503, "application/json", `{}`},
		{"redirect", 307, "application/json", `{}`},
		{"wrong type", 200, "text/plain", `{}`},
		{"oversized", 200, "application/json", strings.Repeat("x", 8193)},
		{"trailing", 200, "application/json", `{} {}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := newAgentReadFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", tc.content)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			if _, err := a.Authorize(t.Context(), agentReadFixtureToken(), agentReadFixtureRequest()); err == nil {
				t.Fatal("dependency failure accepted")
			} else if strings.Contains(err.Error(), agentReadFixtureService) || strings.Contains(err.Error(), agentReadFixtureToken()) {
				t.Fatal("credential leaked in error")
			}
		})
	}
}
