package riidoaiserver

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestDurableACKConcurrentReplicaRetryCommitsOnce(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentRunning)
	a := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	b := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer a.Close()
	defer b.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	acks := make(chan AgentEventResponse, 2)
	errs := make(chan error, 2)
	for _, s := range []*Store{a, b} {
		go func(store *Store) {
			defer wg.Done()
			ack, err := store.RecordAgentEvent(context.Background(), "agent", req)
			acks <- ack
			errs <- err
		}(s)
	}
	wg.Wait()
	close(acks)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	first := <-acks
	second := <-acks
	if !reflect.DeepEqual(first, second) || len(ops.records) != 1 {
		t.Fatal("concurrent identical requests committed twice")
	}
}
