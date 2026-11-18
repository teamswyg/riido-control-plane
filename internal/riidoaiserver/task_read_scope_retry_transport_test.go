package riidoaiserver

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
)

type taskScopeRetryTransport func(*http.Request) (*http.Response, error)

func TestTaskReadScopeRetriesGatewayFailures(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusGatewayTimeout} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client, initial, _ := taskScopeRetryFixture(t, status, 1, 0)
			_, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
			if err != nil || initial.Load() != 2 {
				t.Fatalf("calls=%d err=%v", initial.Load(), err)
			}
		})
	}
}

func TestTaskReadScopeRetriesTimeoutAndInterruptedTransport(t *testing.T) {
	for _, failure := range []error{context.DeadlineExceeded, io.ErrUnexpectedEOF} {
		t.Run(failure.Error(), func(t *testing.T) {
			client, initial, _ := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 0, 0)
			var attempts atomic.Int32
			client.httpClient.Transport = taskScopeRetryTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/graphql" && attempts.Add(1) == 1 {
					return nil, failure
				}
				return http.DefaultTransport.RoundTrip(r)
			})
			_, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
			if err != nil || initial.Load() != 1 || attempts.Load() != 3 {
				t.Fatalf("attempts=%d initial=%d err=%v", attempts.Load(), initial.Load(), err)
			}
		})
	}
}

func (f taskScopeRetryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
