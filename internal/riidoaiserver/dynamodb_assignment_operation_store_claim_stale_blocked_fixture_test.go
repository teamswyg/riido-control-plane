package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type staleBlockedClaimFixture struct {
	Now     time.Time
	Current Assignment
	Blocker Assignment
}

func newStaleBlockedClaimFixture(
	t *testing.T,
) (staleBlockedClaimFixture, *DynamoDBAssignmentOperationStore, <-chan capturedDynamoDBRequest) {
	t.Helper()
	fixture := staleBlockedClaimFixture{
		Now:     time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC),
		Current: staleBlockedCurrentAssignment(),
		Blocker: staleBlockedBlockerAssignment(),
	}
	currentItem := sampleAssignmentProjectionDynamoDBItem(t, fixture.Current, 5)
	blockerItem := sampleAssignmentProjectionDynamoDBItem(t, fixture.Blocker, 3)
	activeItem := staleBlockedActiveItem(fixture.Now)
	getCalls := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixture.Now,
		RequestBuffer: 4,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			switch r.Header.Get("X-Amz-Target") {
			case dynamoDBQueryTarget:
				if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{currentItem}}); err != nil {
					t.Errorf("encode query response: %v", err)
				}
			case dynamoDBGetItemTarget:
				getCalls++
				item := blockerItem
				if getCalls == 2 {
					item = activeItem
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"Item": item}); err != nil {
					t.Errorf("encode get response: %v", err)
				}
			case dynamoDBTransactWriteTarget:
				_, _ = w.Write([]byte(`{}`))
			default:
				t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
				w.WriteHeader(http.StatusBadRequest)
			}
		},
	})
	return fixture, store, requests
}
