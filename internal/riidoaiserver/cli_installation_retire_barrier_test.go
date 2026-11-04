package riidoaiserver

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

type retirementBarrierTransport struct {
	next    http.RoundTripper
	entered chan struct{}
	release chan struct{}
	used    atomic.Bool
}

func (b *retirementBarrierTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget && b.used.CompareAndSwap(false, true) {
		close(b.entered)
		select {
		case <-b.release:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	}
	return b.next.RoundTrip(r)
}

func retirementBarrier(s *DynamoDBAssignmentOperationStore) *retirementBarrierTransport {
	next := s.httpClient.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	b := &retirementBarrierTransport{next: next, entered: make(chan struct{}), release: make(chan struct{})}
	client := *s.httpClient
	client.Transport = b
	s.httpClient = &client
	return b
}

func waitRetirementBarrier(t *testing.T, b *retirementBarrierTransport) {
	t.Helper()
	select {
	case <-b.entered:
	case <-time.After(time.Second):
		t.Fatal("transaction did not reach barrier")
	}
}
