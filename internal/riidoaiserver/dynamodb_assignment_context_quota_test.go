package riidoaiserver

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestDynamoDBAssignmentContextQuotaAtomicAcrossServers(t *testing.T) {
	f, newStore := newQuotaDynamoFixture(t)
	stores := []*DynamoDBAssignmentOperationStore{f.store, newStore()}
	var allowed, rejected, failed atomic.Int32
	var wg sync.WaitGroup
	// Keep simultaneous TCP connects below macOS's local accept queue. We
	// still race 150 total attempts across independent store instances.
	inFlight := make(chan struct{}, 16)
	for i := range 150 {
		wg.Go(func() {
			inFlight <- struct{}{}
			defer func() { <-inFlight }()
			ok, err := stores[i%2].ConsumeAssignmentContextRead(t.Context(), "assignment")
			switch {
			case err != nil:
				t.Logf("quota attempt failed: %v", err)
				failed.Add(1)
			case ok:
				allowed.Add(1)
			default:
				rejected.Add(1)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 100 || rejected.Load() != 50 || failed.Load() != 0 {
		t.Fatalf("allowed=%d rejected=%d failed=%d", allowed.Load(), rejected.Load(), failed.Load())
	}
	if ok, err := newStore().ConsumeAssignmentContextRead(t.Context(), "assignment"); ok || err != nil {
		t.Fatalf("new server replenished budget: allowed=%v err=%v", ok, err)
	}
	if ok, err := newStore().ConsumeAssignmentContextRead(t.Context(), "new-assignment"); !ok || err != nil {
		t.Fatalf("new execution denied: allowed=%v err=%v", ok, err)
	}
}

func TestDynamoDBAssignmentContextQuotaTerminalFencesFirstAndLaterReads(t *testing.T) {
	for _, firstRead := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-first-read", true: "after-first-read"}[firstRead], func(t *testing.T) {
			f, _ := newQuotaDynamoFixture(t)
			if firstRead {
				if ok, err := f.store.ConsumeAssignmentContextRead(t.Context(), "assignment"); !ok || err != nil {
					t.Fatal(ok, err)
				}
			}
			record := AssignmentOperationRecord{AssignmentID: "assignment", Assignment: Assignment{ID: "assignment", State: AssignmentCompleted, UpdatedAt: f.now}, RecordedAt: f.now}
			credentials, _ := f.store.credentialsProvider.Credentials(t.Context())
			if err := f.store.saveAssignmentProjection(t.Context(), record, credentials); err != nil {
				t.Fatal(err)
			}
			if ok, err := f.store.ConsumeAssignmentContextRead(t.Context(), "assignment"); ok || err != nil {
				t.Fatalf("terminal quota must be fenced: allowed=%v err=%v", ok, err)
			}
		})
	}
}
