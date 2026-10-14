package riidoaiserver

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type blockingAuthorizer struct {
	calls   atomic.Int64
	once    sync.Once
	started chan struct{}
	release chan struct{}
}

func newBlockingAuthorizer() *blockingAuthorizer {
	return &blockingAuthorizer{started: make(chan struct{}), release: make(chan struct{})}
}

func (a *blockingAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	a.calls.Add(1)
	a.once.Do(func() { close(a.started) })
	<-a.release
	return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-1"}, nil
}

func runConcurrentAuthorization(t *testing.T, authorizer RequestAuthorizer, token string, req AuthorizationRequest, count int, next *blockingAuthorizer) []AuthorizationResult {
	t.Helper()
	start := make(chan struct{})
	results := make([]AuthorizationResult, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			result, err := authorizer.Authorize(context.Background(), token, req)
			if err != nil {
				t.Errorf("Authorize: %v", err)
			}
			results[i] = result
		}(i)
	}
	close(start)
	<-next.started
	time.Sleep(20 * time.Millisecond)
	close(next.release)
	wg.Wait()
	return results
}

func (a *blockingAuthorizer) waitForCalls(t *testing.T, want int64) {
	t.Helper()
	deadline := time.After(time.Second)
	for a.calls.Load() < want {
		select {
		case <-deadline:
			t.Fatalf("underlying calls = %d, want %d", a.calls.Load(), want)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
