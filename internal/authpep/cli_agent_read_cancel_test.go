package authpep

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestCLIAgentReadCancellationCannotReturnAuthority(t *testing.T) {
	entered, stopped := make(chan struct{}), make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	a, _ := newAgentReadFixture(t, func(_ http.ResponseWriter, r *http.Request) {
		assertAgentReadFixtureRequest(t, r)
		close(entered)
		select {
		case <-r.Context().Done():
			close(stopped)
		case <-release:
		}
	})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		result, err := a.Authorize(ctx, agentReadFixtureToken(), agentReadFixtureRequest())
		if result.PrincipalID != "" {
			t.Error("cancelled request returned authority")
		}
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("authorization did not reach service")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled authorization succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation did not end authorization")
	}
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream request was not cancelled")
	}
}
