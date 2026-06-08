package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Work already queued -> WaitForAssignment returns immediately, no hold.
func TestWaitForAssignmentReturnsImmediatelyWhenQueued(t *testing.T) {
	store := NewStore()
	defer store.Close()
	if _, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-a", RuntimeProvider: "codex", Prompt: "go",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}

	start := time.Now()
	resp, err := store.WaitForAssignment(context.Background(), "agent-a",
		PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 5000}, 5*time.Second, time.Second)
	if err != nil {
		t.Fatalf("WaitForAssignment: %v", err)
	}
	if resp.Action != PollStart {
		t.Fatalf("action = %s, want start", resp.Action)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("returned after %v; should be immediate when work is already queued", elapsed)
	}
}

// No work, then an assignment is created during the hold -> the waiter is
// signalled and returns start well before the budget elapses.
func TestWaitForAssignmentWakesOnAssign(t *testing.T) {
	store := NewStore()
	defer store.Close()

	type result struct {
		resp PollResponse
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := store.WaitForAssignment(context.Background(), "agent-a",
			PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 5000}, 5*time.Second, 4*time.Second)
		done <- result{resp, err}
	}()

	// Give the waiter time to register, then queue work.
	time.Sleep(50 * time.Millisecond)
	if _, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-a", RuntimeProvider: "codex", Prompt: "go",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("WaitForAssignment: %v", res.err)
		}
		if res.resp.Action != PollStart {
			t.Fatalf("action = %s, want start", res.resp.Action)
		}
	case <-time.After(2 * time.Second):
		// 2s << the 4s tick, so success here proves the signal woke it, not the
		// fallback re-poll tick.
		t.Fatal("WaitForAssignment did not wake on assign signal")
	}
}

// No work and nothing queued -> returns action=none after the hold budget.
func TestWaitForAssignmentTimesOutWithNone(t *testing.T) {
	store := NewStore()
	defer store.Close()

	start := time.Now()
	resp, err := store.WaitForAssignment(context.Background(), "agent-a",
		PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 200}, 200*time.Millisecond, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForAssignment: %v", err)
	}
	if resp.Action != PollNone {
		t.Fatalf("action = %s, want none", resp.Action)
	}
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Fatalf("returned after %v; should hold ~200ms before timing out", elapsed)
	}

	// The whole held poll counts as exactly one daemon poll request, not one per
	// internal re-evaluation tick.
	snapshot, err := store.Metrics(context.Background())
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if snapshot.PollRequestsTotal != 1 {
		t.Fatalf("PollRequestsTotal = %d, want 1 (held poll must count once)", snapshot.PollRequestsTotal)
	}
}

// ctx cancellation during a hold returns promptly with the ctx error.
func TestWaitForAssignmentCtxCancel(t *testing.T) {
	store := NewStore()
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := store.WaitForAssignment(ctx, "agent-a",
			PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 5000}, 5*time.Second, time.Second)
		done <- err
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("WaitForAssignment should return ctx error on cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForAssignment did not return promptly after ctx cancel")
	}
}

// The actor loop keeps serving other commands while a poll is held (the wait
// runs off the actor goroutine).
func TestWaitForAssignmentDoesNotBlockActor(t *testing.T) {
	store := NewStore()
	defer store.Close()

	go func() {
		_, _ = store.WaitForAssignment(context.Background(), "agent-held",
			PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 3000}, 3*time.Second, time.Second)
	}()
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	if _, err := store.Metrics(context.Background()); err != nil {
		t.Fatalf("Metrics during held poll: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Metrics took %v during a held long-poll; actor appears blocked", elapsed)
	}
}

// HTTP handler: a wait_ms request takes the long-poll branch, clamps the hold to
// the server budget, and returns 200 + action=none on timeout.
func TestHTTPAgentPollLongPollHoldsClampsAndTimesOut(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{
		Assignment:      store,
		Authorizer:      assignmentHTTPAuthorizer(t, []string{"agent:agent-a:poll"}),
		LongPollMaxHold: 150 * time.Millisecond,
		LongPollTick:    40 * time.Millisecond,
	}).Handler()

	// wait_ms (5000) far exceeds the 150ms server cap: clamp -> ~150ms hold.
	body := `{"daemon_id":"daemon-a","device_id":"device-a","runtime_id":"runtime-a","wait_ms":5000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	start := time.Now()
	server.ServeHTTP(resp, req)
	elapsed := time.Since(start)

	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out PollResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if out.Action != PollNone {
		t.Fatalf("action=%s, want none", out.Action)
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("held %v; expected ~150ms hold (long-poll branch)", elapsed)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("held %v; clamp to LongPollMaxHold did not apply", elapsed)
	}
}
