package riidoaiserver

import (
	"net/http/httptest"
	"sync"
	"testing"
)

func TestExecutionTerminationConcurrentReplicasPreserveOneReceipt(t *testing.T) {
	f, newStore, a := newTerminationDynamoFixture(t)
	handlers := []*terminationHTTPFixture{terminationDynamoHTTP(t, newStore(), a), terminationDynamoHTTP(t, newStore(), a)}
	results := make(chan *httptest.ResponseRecorder, 8)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h := handlers[i%2]
			results <- serveTermination(h.handler, h.request(terminationFixtureBody))
		}()
	}
	wg.Wait()
	close(results)
	first := ""
	for result := range results {
		if result.Code != 200 {
			t.Fatalf("concurrent status=%d", result.Code)
		}
		if first == "" {
			first = result.Body.String()
		}
		if first != result.Body.String() {
			t.Fatal("concurrent retries changed receipt")
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.commits != 1 {
		t.Fatalf("committed receipts=%d", f.commits)
	}
}

func TestExecutionTerminationDoesNotPublishTaskEvents(t *testing.T) {
	f := newTerminationHTTPFixture(t, StoreConfig{})
	_, events, cancel, err := f.actor.SubscribeTask(t.Context(), "task")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	if result := serveTermination(f.handler, f.request(terminationFixtureBody)); result.Code != 200 {
		t.Fatalf("status=%d", result.Code)
	}
	select {
	case <-events:
		t.Fatal("private receipt entered public task events")
	default:
	}
}
