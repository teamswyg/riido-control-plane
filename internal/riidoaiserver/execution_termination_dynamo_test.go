package riidoaiserver

import (
	"crypto/sha256"
	"strings"
	"testing"
)

func terminationDynamoHTTP(t *testing.T, store *DynamoDBAssignmentOperationStore, a Assignment) *terminationHTTPFixture {
	t.Helper()
	actor := NewStoreWithConfig(StoreConfig{OperationStore: store})
	t.Cleanup(actor.Close)
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["device"] = deviceCredentialRecord{deviceID: "device", secretHash: sha256.Sum256([]byte("secret-sentinel")), ownerPrincipalID: "owner"}
	return &terminationHTTPFixture{actor, devices, NewServer(ServerConfig{Assignment: actor, AIAgentClient: devices}).Handler(), a}
}

func TestExecutionTerminationHTTPCommitsClaimBoundDynamoReceipt(t *testing.T) {
	f, newStore, a := newTerminationDynamoFixture(t)
	h := terminationDynamoHTTP(t, newStore(), a)
	first := serveTermination(h.handler, h.request(terminationFixtureBody))
	if first.Code != 200 {
		t.Fatalf("status=%d", first.Code)
	}
	f.mu.Lock()
	if len(f.actions) != 4 || f.actions[0].ConditionCheck == nil || f.actions[1].ConditionCheck == nil || f.actions[2].Put == nil || !isDevicePolicyAdvance(f.actions[3]) {
		t.Error("claim, projection and receipt must commit atomically")
	}
	for _, body := range f.requests {
		if strings.Contains(string(body), "lease-sentinel") || strings.Contains(string(body), "secret-sentinel") {
			t.Error("storage request leaked credential")
		}
	}
	f.mu.Unlock()
	reopened := terminationDynamoHTTP(t, newStore(), a)
	again := serveTermination(reopened.handler, reopened.request(terminationFixtureBody))
	if again.Code != 200 || again.Body.String() != first.Body.String() {
		t.Fatal("restart lost immutable receipt")
	}
	w := serveTermination(reopened.handler, reopened.request(strings.Replace(terminationFixtureBody, `:2}`, `:3}`, 1)))
	if w.Code != 409 {
		t.Fatalf("changed payload status=%d", w.Code)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.commits != 1 {
		t.Fatalf("receipt commits=%d", f.commits)
	}
}
