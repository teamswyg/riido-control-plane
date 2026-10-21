package riidoaiserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalescingAuthorizerKeepsExternalTimeoutAndSingleAttempt(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	next, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: server.URL, Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := NewCoalescingAuthorizer(next)
	result, err := a.Authorize(context.Background(), "token", AuthorizationRequest{Resource: AuthorizationResourceMetrics})
	if !errors.Is(err, context.DeadlineExceeded) || result.PrincipalID != "" || calls.Load() != 1 {
		t.Fatalf("timeout lost or authorization retried: err=%v calls=%d", err, calls.Load())
	}
}
