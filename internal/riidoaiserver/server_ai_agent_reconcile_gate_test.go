package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentGlobalReconcileGateReserveForgetAndPrune(t *testing.T) {
	gate := newAIAgentGlobalReconcileGate(time.Minute)
	principal := AuthorizationResult{WorkspaceID: " workspace ", PrincipalID: " user "}
	now := time.Date(2026, 7, 7, 1, 2, 3, 0, time.UTC)

	key, ok := gate.reserve(principal, now)
	if !ok || key != "workspace\x00user" {
		t.Fatalf("first reserve = key %q ok %v", key, ok)
	}
	if _, ok := gate.reserve(principal, now.Add(30*time.Second)); ok {
		t.Fatal("second reserve inside interval should be blocked")
	}
	gate.forget(key)
	if _, ok := gate.reserve(principal, now.Add(31*time.Second)); !ok {
		t.Fatal("reserve after forget should be allowed")
	}
	if _, ok := gate.reserve(principal, now.Add(4*time.Minute)); !ok {
		t.Fatal("reserve after prune window should be allowed")
	}
}

func TestAIAgentGlobalReconcileGateNilAndDefaultInterval(t *testing.T) {
	var gate *aiAgentGlobalReconcileGate
	if _, ok := gate.reserve(AuthorizationResult{}, time.Now()); !ok {
		t.Fatal("nil gate should allow reserve")
	}
	gate.forget("ignored")

	defaulted := newAIAgentGlobalReconcileGate(0)
	if defaulted.interval != time.Second {
		t.Fatalf("default interval = %v, want 1s", defaulted.interval)
	}
	defaulted.forget(" ")
}
