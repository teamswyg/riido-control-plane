package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestContextReadReplaySharesOneBudgetAndKeepsTerminal(t *testing.T) {
	reads, queries := make([]ClientStreamEvent, 0, 57), 0
	for i := range 56 {
		reads = append(reads, ClientStreamEvent{Seq: int64(i + 1), EventType: AgentClientEventContextRead})
	}
	reads = append(reads, ClientStreamEvent{Seq: 57, EventType: AgentClientEventWorkStatusChanged})
	filter := func(ctx context.Context, event ClientStreamEvent) (ClientStreamEvent, bool) {
		if event.EventType != AgentClientEventContextRead {
			return event, true
		}
		if ctx.Err() == nil {
			queries++
			<-ctx.Done()
		}
		return ClientStreamEvent{}, false
	}
	w := httptest.NewRecorder()
	started := time.Now()
	serveAIAgentClientEventStream(w, httptest.NewRequest(http.MethodGet, "/events?replay=1", nil), reads, nil, filter)
	if queries != 1 || time.Since(started) > DefaultAIAgentTaskContextClientTimeout+time.Second || !strings.Contains(w.Body.String(), "id: 57\n") {
		t.Fatalf("replay budget accumulated or terminal lost: queries=%d body=%s", queries, w.Body.String())
	}
}

func TestContextReadDeliveryDoesNotWaitForBusyStore(t *testing.T) {
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	store.operationMu.Lock()
	defer store.operationMu.Unlock()
	event := ClientStreamEvent{EventType: AgentClientEventContextRead, Payload: AgentContextReadEvent{AgentID: "agent-owned-codex"}}
	if _, allowed := store.FilterAIAgentClientEvent(t.Context(), AuthorizationResult{PrincipalID: "user-1"}, event); allowed {
		t.Fatal("busy persistent store must drop this hint")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, allowed := store.DevelopmentAIAgentClientStore.FilterAIAgentClientEvent(t.Context(), AuthorizationResult{PrincipalID: "user-1"}, event); allowed {
		t.Fatal("busy local store must drop this hint")
	}
}
