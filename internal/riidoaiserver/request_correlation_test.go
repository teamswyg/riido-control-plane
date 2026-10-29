package riidoaiserver

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestCorrelationIgnoresCallerAndSurvivesAuthorizationFailure(t *testing.T) {
	var ids []string
	handler := withRequestCorrelation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids = append(ids, requestCorrelationID(r))
		w.WriteHeader(http.StatusUnauthorized)
	}))
	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", nil)
		req.Header.Set("X-Riido-Request-Id", "caller-secret")
		out := httptest.NewRecorder()
		handler.ServeHTTP(out, req)
		id := out.Header().Get("X-Riido-Request-Id")
		if decoded, err := hex.DecodeString(id); err != nil || len(decoded) != 16 || ids[len(ids)-1] != id {
			t.Fatalf("invalid correlation: %q", id)
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("requests reused correlation")
	}
}
