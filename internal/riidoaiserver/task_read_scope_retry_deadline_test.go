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

func TestTaskReadScopeSlowInitialReadCannotResetDeadline(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, Timeout: 30 * time.Millisecond})
	var out taskReadScope
	err := client.readScopeGraphQL(t.Context(), "synthetic", readScopeQuery, "getComponentReadScope", nil, &out)
	if !errors.Is(err, context.DeadlineExceeded) || calls.Load() != 1 {
		t.Fatalf("budget reset: calls=%d err=%v", calls.Load(), err)
	}
}

func TestTaskReadScopeCancellationNeverMasksPermanentFailure(t *testing.T) {
	for _, outcome := range []string{"denied", "missing", "invalid", "integrity"} {
		ctx, cancel := context.WithCancel(t.Context())
		err := retryTaskContextRead(ctx, func() error {
			cancel()
			return taskContextReadError{outcome}
		})
		var failure taskContextReadError
		if !errors.As(err, &failure) || failure.outcome != outcome || readScopeIncomplete(ctx, err) {
			t.Fatalf("%s hidden by cancellation: %v", outcome, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := retryTaskContextRead(ctx, func() error { t.Fatal("read after cancellation"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read: %v", err)
	}
}
