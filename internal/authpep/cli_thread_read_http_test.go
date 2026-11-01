package authpep

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCLIThreadReadHTTPAlwaysChecksCurrentGrant(t *testing.T) {
	var calls atomic.Int64
	a, _ := newThreadReadFixture(t, func(w http.ResponseWriter, r *http.Request) {
		assertThreadReadFixtureRequest(t, r)
		if calls.Add(1) == 3 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(threadReadFixtureResponse())
	})
	for range 2 {
		result, err := a.Authorize(t.Context(), threadReadFixtureToken(), threadReadFixtureRequest())
		if err != nil || result.PrincipalID != "user-1" || result.WorkspaceID != threadReadFixtureWorkspace {
			t.Fatal("valid live check rejected", err)
		}
	}
	if _, err := a.Authorize(t.Context(), threadReadFixtureToken(), threadReadFixtureRequest()); !errors.Is(err, ErrAuthorizationForbidden) || calls.Load() != 3 {
		t.Fatal("revocation was cached", err, calls.Load())
	}
}

func TestCLIThreadReadHTTPRejectsDependencyFailures(t *testing.T) {
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
			a, _ := newThreadReadFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", tc.content)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			if _, err := a.Authorize(t.Context(), threadReadFixtureToken(), threadReadFixtureRequest()); err == nil {
				t.Fatal("dependency failure accepted")
			} else if strings.Contains(err.Error(), threadReadFixtureService) || strings.Contains(err.Error(), threadReadFixtureToken()) {
				t.Fatal("credential leaked in error")
			}
		})
	}
}
