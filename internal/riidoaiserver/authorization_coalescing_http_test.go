package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSharedAuthorizationSurvivesFirstClientDisconnect(t *testing.T) {
	started, release, joined := make(chan struct{}), make(chan struct{}), make(chan struct{})
	next := authorizationFunc(func(ctx context.Context, _ string, _ AuthorizationRequest) (AuthorizationResult, error) {
		close(started)
		select {
		case <-ctx.Done():
			return AuthorizationResult{}, ctx.Err()
		case <-release:
			return AuthorizationResult{PrincipalID: "requester"}, nil
		}
	})
	a, _ := NewCoalescingAuthorizer(next)
	s := Server{config: ServerConfig{Authorizer: a}}
	ownerHandlerDone := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Waiter") == "true" {
			r = r.WithContext(&joinedAuthorizationContext{Context: r.Context(), joined: joined})
		} else {
			defer close(ownerHandlerDone)
		}
		if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead}); ok {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	first.Header.Set(aiAgentTokenHeader, "same-token")
	firstDone := make(chan struct{})
	go func() {
		response, _ := server.Client().Do(first)
		if response != nil {
			_ = response.Body.Close()
		}
		close(firstDone)
	}()
	<-started
	second, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	second.Header.Set(aiAgentTokenHeader, "same-token")
	second.Header.Set("X-Test-Waiter", "true")
	status := make(chan int, 1)
	go func() {
		response, err := server.Client().Do(second)
		if err != nil {
			status <- 0
			return
		}
		defer response.Body.Close()
		status <- response.StatusCode
	}()
	<-joined
	cancel()
	<-firstDone
	<-ownerHandlerDone
	close(release)
	if got := <-status; got != http.StatusOK {
		t.Fatalf("live HTTP caller inherited first client disconnect: %d", got)
	}
}
